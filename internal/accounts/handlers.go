package accounts

import (
	"fmt"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
)

func RegisterHandlers(authRoutes *gin.RouterGroup, publicRoutes *gin.RouterGroup) {
	publicRoutes.POST("/register", handleRegister)

	authRoutes.POST("/completeregister", handleCompleteRegister)
	authRoutes.POST("/registertoken", CompleteRegisteredMiddleware, handleRegisterToken)
	authRoutes.POST("/updateavatar", CompleteRegisteredMiddleware, handleUpdateAvatar)

	authRoutes.GET("/users/id/:uid", CompleteRegisteredMiddleware, handleGetUser)
	authRoutes.GET("/uid/:username", CompleteRegisteredMiddleware, handleGetUid)

}

func CreateDBUserRecords(ctx *gin.Context, uid string, username string, email string) bool {
	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	if err := dbCreateUserRecordsInternal(ctx, mongoInst, uid, username, email); err != nil {
		comm.AbortBadRequest(ctx, err.Error(), comm.CodeCantCreateAuthUser)
		return false
	}

	return true
}
func handleGetUser(ctx *gin.Context) {
	targetUserId := ctx.Param("uid")

	userData := UserData{}
	if !DBGetUserData(ctx, targetUserId, &userData) {
		return
	}

	comm.GenericOKJSON(ctx, userData)
}
func handleGetUid(ctx *gin.Context) {
	targetUsername := ctx.Param("username")

	usernameData := UsernameData{}
	if !DBGetUsernameData(ctx, targetUsername, &usernameData) {
		return
	}

	comm.GenericOKJSON(ctx, usernameData)
}

// TODO: remove trailing spaces and check for correct username format
type RegisterParams struct {
	Username string `json:"username" binding:"min=4,alphanum,required"`
	Email    string `json:"email" binding:"email,required"`
	Pwd      string `json:"pwd" binding:"min=6,required"`
}

func handleRegister(ctx *gin.Context) {
	params := RegisterParams{}

	if err := ctx.ShouldBind(&params); err != nil {
		comm.AbortFailedBinding(ctx, err)

		return
	}

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	usernamesCollection := mongoInst.Collection(database.UsernamesCollection)

	filter := bson.D{{Key: "_id", Value: params.Username}}
	usernameRes := usernamesCollection.FindOne(ctx, filter)
	if usernameRes.Err() == nil {
		comm.AbortBadRequest(ctx, "User already exists", comm.CodeUsernameTaken)
		return
	}

	fbAuth, _ := ctx.MustGet(auth.CtxVarFirebaseAuth).(auth.Auth)

	userCreateParams := (&fbauth.UserToCreate{}).
		Email(params.Email).
		EmailVerified(false).
		Password(params.Pwd).
		Disabled(false)

	userRecord, err := fbAuth.CreateUser(ctx, userCreateParams)
	if err != nil {
		respMsg := fmt.Sprintf("Backend failed to create new user with %s", err.Error())
		comm.AbortBadRequest(ctx, respMsg, comm.CodeCantCreateAuthUser)
		return
	}

	if !CreateDBUserRecords(ctx, userRecord.UID, params.Username, params.Email) {
		return
	}

	comm.OK(ctx, "Registered", comm.CodeSuccess)
}

type UpdateAvatarParams struct {
	ImgUrl string `json:"img_url" binding:"gte=1,lte=1024,url,required"`
}

func handleUpdateAvatar(ctx *gin.Context) {
	params := UpdateAvatarParams{}
	err := ctx.ShouldBind(&params)
	if err != nil {
		comm.AbortFailedBinding(ctx, err)
		return
	}

	userId := ctx.MustGet(auth.CtxVarUserId).(string)
	if len(userId) == 0 {
		comm.AbortUnauthorized(ctx, "Invalid creds", comm.CodeNotAuthenticated)
		return //not authenticated
	}
	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

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
	Username string `json:"username" binding:"min=4,alphanum,required"`
}

func handleCompleteRegister(ctx *gin.Context) {
	params := CompleteRegisterParams{}
	if err := ctx.ShouldBind(&params); err != nil {
		comm.AbortFailedBinding(ctx, err)
		return
	}

	userId := ctx.MustGet(auth.CtxVarUserId).(string)
	if len(userId) == 0 {
		comm.AbortUnauthorized(ctx, "Invalid creds", comm.CodeNotAuthenticated)
		return // not authenticated
	}

	authToken := ctx.MustGet(auth.CtxVarAuthToken).(*fbauth.Token)
	auth := ctx.MustGet(auth.CtxVarFirebaseAuth).(auth.Auth)
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

	if !CreateDBUserRecords(ctx, authToken.UID, params.Username, email) {
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
	userId := ctx.MustGet(auth.CtxVarUserId).(string)
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

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)
	usersCollection := mongoInst.Collection(database.UsersCollection)

	filter := bson.D{{Key: "_id", Value: userId}}
	res := usersCollection.FindOne(ctx, filter)
	if res.Err() != nil {
		comm.AbortBadRequest(ctx, "Auth error", comm.CodeNotAuthenticated)
		return
	}

	userData := UserData{}
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
