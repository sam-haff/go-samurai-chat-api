package auth

import (
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/middleware"
	"go-chat-app-api/internal/users"
)

func RegisterHandlers(routers *gin.Engine) {
	routers.POST("/register", handleRegister)
	routers.POST("/completeregister", middleware.AuthMiddleware, handleCompleteRegister)
	routers.POST("/registertoken", middleware.AuthMiddleware, handleRegisterToken)
	routers.POST("/updateavatar", middleware.AuthMiddleware, handleUpdateAvatar)
}

func createDBUserRecords(ctx *gin.Context, uid string, username string, email string) bool {
	mongoInst := ctx.MustGet(middleware.CtxVarMongoDBInst).(*database.MongoDBInstance)
	usersCollection := mongoInst.Collection(database.UsersCollection)
	usernamesCollection := mongoInst.Collection(database.UsernamesCollection)

	wc := writeconcern.Majority()
	txOptions := options.Transaction().SetWriteConcern(wc)
	session, err := mongoInst.Client.StartSession()
	if err != nil {
		fmt.Printf("Failed to start session \n")
		comm.AbortBadRequest(ctx, "Failed to start tx", comm.CodeCantCreateAuthUser)
		return false
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx,
		func(ctx mongo.SessionContext) (interface{}, error) {
			_, err := usersCollection.InsertOne(ctx, users.UserData{
				Id:       uid,
				Username: username,
				Email:    email,
			})
			if err != nil {
				return nil, err
			}
			_, err = usernamesCollection.InsertOne(ctx, users.UsernameData{
				Id:     username,
				UserID: uid,
			})

			return nil, err
		},
		txOptions,
	)
	if err != nil {

		fmt.Printf("Tx failed with error: %s\n", err.Error())
		comm.AbortBadRequest(ctx, "Failed to create db records", comm.CodeCantCreateAuthUser)
		return false
	}
	return true
}

type RegisterParams struct {
	Username string `json:"username" binding:"min=4,required"`
	Email    string `json:"email" binding:"email,required"`
	Pwd      string `json:"pwd" binding:"min=6,required"`
}

func handleRegister(ctx *gin.Context) {
	params := RegisterParams{}

	if err := ctx.ShouldBind(&params); err != nil {
		comm.AbortFailedBinding(ctx, err)

		return
	}

	fbApp := ctx.MustGet(middleware.CtxVarFirebaseApp).(*firebase.App)
	mongoInst := ctx.MustGet(middleware.CtxVarMongoDBInst).(*database.MongoDBInstance)

	usernamesCollection := mongoInst.Collection(database.UsernamesCollection)

	filter := bson.D{{Key: "_id", Value: params.Username}}
	usernameRes := usernamesCollection.FindOne(ctx, filter)
	if usernameRes.Err() == nil {
		comm.AbortBadRequest(ctx, "User already exists", comm.CodeUsernameTaken)
		return
	}

	if strings.Contains(params.Username, " ") || len(params.Username) < 4 {
		comm.AbortBadRequest(ctx, "Invalid username", comm.CodeUsernameFormatNotValid)
		return
	}

	fbAuth, _ := fbApp.Auth(ctx)

	userCreateParams := (&auth.UserToCreate{}).
		Email(params.Email).
		EmailVerified(false).
		Password(params.Pwd).
		Disabled(false)

	userRecord, err := fbAuth.CreateUser(ctx, userCreateParams)
	if err != nil {
		respMsg := fmt.Sprintf("Backend failed to create new user with %s", err.Error())
		comm.AbortBadRequest(ctx, respMsg, comm.CodeCantCreateAuthUser)
		//ctx.String(400, apiResponse(fmt.Sprintf("Backend failed to create new user with %s", err.Error()), CodeCantCreateAuthUser))
		return
	}

	if !createDBUserRecords(ctx, userRecord.UID, params.Username, params.Email) {
		return
	}

	comm.OK(ctx, "Registered", comm.CodeSuccess)
}

type UpdateAvatarParams struct {
	ImgUrl string `json:"img_url" binding:"gte=1,required"`
}

func handleUpdateAvatar(ctx *gin.Context) {
	params := UpdateAvatarParams{}
	err := ctx.ShouldBind(&params)
	if err != nil {
		comm.AbortFailedBinding(ctx, err)
		return
	}

	userId := ctx.MustGet(middleware.CtxVarUserId).(string)
	if len(userId) == 0 {
		comm.AbortUnauthorized(ctx, "Invalid creds", comm.CodeNotAuthenticated)
		return //not authenticated
	}
	mongoInst := ctx.MustGet(middleware.CtxVarMongoDBInst).(*database.MongoDBInstance)

	filter := bson.D{{Key: "_id", Value: userId}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "img_url", Value: params.ImgUrl}}}}
	_, err = mongoInst.Collection(database.UsersCollection).UpdateOne(ctx, filter, update)

	if err != nil {
		comm.AbortBadRequest(ctx, "Failed to update url", comm.CodeInvalidArgs)
		return
	}
	comm.GenericOK(ctx)
}

type CompleteRegisterParams struct {
	Username string `json:"username" binding:"min=4,required"`
}

func handleCompleteRegister(ctx *gin.Context) {
	params := CompleteRegisterParams{}
	if err := ctx.ShouldBind(&params); err != nil {
		comm.AbortFailedBinding(ctx, err)
		return
	}

	userId := ctx.MustGet(middleware.CtxVarUserId).(string)
	if len(userId) == 0 {
		comm.AbortUnauthorized(ctx, "Invalid creds", comm.CodeNotAuthenticated)
		return // not authenticated
	}

	authToken := ctx.MustGet(middleware.CtxVarAuthToken).(*auth.Token)
	fbApp := ctx.MustGet(middleware.CtxVarFirebaseApp).(*firebase.App)
	auth, _ := fbApp.Auth(ctx)
	userRecord, err := auth.GetUser(ctx, authToken.UID)
	fmt.Printf(userRecord.ProviderID + "\n")
	fmt.Printf(authToken.Firebase.SignInProvider + "\n")
	if err != nil {
		comm.AbortUnauthorized(ctx, "Smth went wrong", comm.CodeNotAuthenticated)
		return
	}
	email := ""
	for _, v := range userRecord.ProviderUserInfo {
		if v.ProviderID == authToken.Firebase.SignInProvider {
			email = v.Email
			break
		}
	}
	if len(email) == 0 {
		fmt.Printf("Invalid provider\n")
		comm.AbortUnauthorized(ctx, "Invalid provider", comm.CodeNotAuthenticated)
		return
	}

	if !createDBUserRecords(ctx, authToken.UID, params.Username, email) {
		//all responces are handled inside func
		return
	}

	comm.OK(ctx, "Completed registration", comm.CodeSuccess)
}

type RegisterTokenParams struct {
	Token      string `json:"token" binding:"gte=1,lte=2049,required"`
	DeviceName string `json:"device_name" binding:"gte=1,lte=2049,required"`
}

func handleRegisterToken(ctx *gin.Context) {
	userId := ctx.MustGet(middleware.CtxVarUserId).(string)
	if len(userId) == 0 {
		comm.AbortUnauthorized(ctx, "Invalid creds", comm.CodeNotAuthenticated)
		return // not authenticated
	}
	params := RegisterTokenParams{}
	err := ctx.ShouldBind(&params)
	if err != nil {
		fmt.Printf("Couldnt bind register token params: %s\n", err.Error())

		comm.AbortFailedBinding(ctx, err)
		return
	}

	mongoInst := ctx.MustGet(middleware.CtxVarMongoDBInst).(*database.MongoDBInstance)
	usersCollection := mongoInst.Collection(database.UsersCollection)

	filter := bson.D{{Key: "_id", Value: userId}}
	res := usersCollection.FindOne(ctx, filter)
	if res.Err() != nil {
		comm.AbortBadRequest(ctx, "Auth error", comm.CodeNotAuthenticated)
		return
	}

	userData := users.UserData{}
	err = res.Decode(&userData)
	if err != nil {
		comm.AbortBadRequest(ctx, "Failed to device data from db", comm.CodeInvalidArgs)
		return
	}

	if userData.Tokens == nil {
		userData.Tokens = make(map[string]string)
	}
	userData.Tokens[params.DeviceName] = params.Token

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "tokens", Value: userData.Tokens}}}} //bson.D{$set: {"tokens", userData.Tokens}}
	_, err = usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		respMsg := fmt.Sprintf("Failed to write tokens to db with: %s", err.Error())

		comm.AbortBadRequest(ctx, respMsg, comm.CodeInvalidArgs)
		return
	}

	comm.OK(ctx, "Token registered", comm.CodeSuccess)
}
