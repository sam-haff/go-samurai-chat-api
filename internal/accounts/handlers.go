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

func RegisterUploadHandlers(authRoutes *gin.RouterGroup, publicRoutes *gin.RouterGroup) {
}
func RegisterHandlers(authRoutes *gin.RouterGroup, publicRoutes *gin.RouterGroup) {
	publicRoutes.POST("/register", handleRegister)

	authRoutes.POST("/register/complete", handleCompleteRegister)
	authRoutes.POST("/registertoken", CompleteRegisteredMiddleware, handleRegisterToken)
	authRoutes.POST("/addcontact", CompleteRegisteredMiddleware, handleAddContact) // TODO: use dynamic path parameter?
	authRoutes.POST("/removecontact", CompleteRegisteredMiddleware, handleRemoveContact)

	authRoutes.GET("/users/id/:uid", CompleteRegisteredMiddleware, handleGetUser)
	authRoutes.GET("/users/username/:username", CompleteRegisteredMiddleware, handleGetUserByUsername)
}

func CreateDBUserRecords(ctx *gin.Context, userData UserData) bool {
	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	if err := dbCreateUserRecordsInternal(ctx, mongoInst, userData); err != nil {
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
func handleGetUserByUsername(ctx *gin.Context) {
	targetUsername := ctx.Param("username")

	userData := UserData{}
	if !DBGetUserDataByUsername(ctx, targetUsername, &userData) {
		return
	}

	comm.GenericOKJSON(ctx, userData)
}

type ContactParams struct {
	// TODO: add binding rules
	Username string `json:"username" binding:"min=4,alphanum,required"`
}

func handleAddContact(ctx *gin.Context) {
	params := ContactParams{}
	if err := ctx.ShouldBind(&params); err != nil {
		comm.AbortFailedBinding(ctx, err)

		return
	}

	user := ctx.MustGet(CtxVarUserData).(UserData)

	contact := UserData{}
	if !DBGetUserDataByUsername(ctx, params.Username, &contact) {
		return
	}

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)
	res := mongoInst.Collection(database.UsersCollection).FindOneAndUpdate(
		ctx,
		bson.M{"_id": user.Id},
		bson.M{"$set": bson.M{"contacts." + contact.Id: true}},
	)
	old := UserData{}
	err := res.Decode(&old)
	if err != nil {
		comm.AbortBadRequest(ctx, "Database failure", comm.CodeInvalidArgs)
		return
	}

	_, ok := old.Contacts[contact.Id]
	if ok {
		comm.AbortBadRequest(ctx, "Contact is already in the list", comm.CodeInvalidArgs)
		return
	}

	comm.GenericOKJSON(ctx, contact)
}

func handleRemoveContact(ctx *gin.Context) {
	params := ContactParams{}
	if err := ctx.ShouldBind(&params); err != nil {
		comm.AbortFailedBinding(ctx, err)

		return
	}

	user := ctx.MustGet(CtxVarUserData).(UserData)

	contact := UserData{}
	if !DBGetUserDataByUsername(ctx, params.Username, &contact) {
		return
	}

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)
	res := mongoInst.Collection(database.UsersCollection).FindOneAndUpdate(
		ctx,
		bson.M{"_id": user.Id},
		bson.M{"$unset": bson.M{"contacts." + contact.Id: true}},
	)
	old := UserData{}
	err := res.Decode(&old)
	if err != nil {
		comm.AbortBadRequest(ctx, "Database failure", comm.CodeInvalidArgs)
		return
	}

	_, ok := old.Contacts[contact.Id]
	if !ok {
		comm.AbortBadRequest(ctx, "Contact is not in the list", comm.CodeInvalidArgs)
		return
	}

	comm.GenericOKJSON(ctx, contact)
}

// TODO: remove trailing spaces and check for correct username format
type RegisterParams struct {
	Username string `json:"username" binding:"min=4,alphanum,required"`
	Email    string `json:"email" binding:"email,required"`
	Pwd      string `json:"pwd" binding:"min=6,required"`
}

func handleRegister(ctx *gin.Context) {
	//TODO: add middleware to check that Firebase Auth record exist
	params := RegisterParams{}

	if err := ctx.ShouldBind(&params); err != nil {
		comm.AbortFailedBinding(ctx, err)

		return
	}

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	// check if already registered
	usersCollection := mongoInst.Collection(database.UsersCollection)
	filter := bson.D{{Key: "username", Value: params.Username}}
	usernameRes := usersCollection.FindOne(ctx, filter)
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

	if !CreateDBUserRecords(ctx, NewUserData(userRecord.UID, params.Email, params.Username, "")) {
		return
	}

	comm.OK(ctx, "Registered", comm.CodeSuccess)
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

	if !CreateDBUserRecords(ctx, NewUserData(authToken.UID, email, params.Username, "")) {
		//all responces are handled inside func
		return
	}

	comm.OK(ctx, "Completed registration", comm.CodeSuccess)
}

const (
	MaxFcmTokenLength      = 2049
	MaxFcmDeviceNameLength = 2049
)

type RegisterTokenParams struct {
	Token      string `json:"token" binding:"min=1,max=2049,required"`
	DeviceName string `json:"device_name" binding:"min=1,max=2049,required"`
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
	//TODO: check on current keys count so that the number is not too big(<=32 for example)
	userData.Tokens[params.DeviceName] = params.Token
	//TODO: bad
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "tokens", Value: userData.Tokens}}}}
	_, err = usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		respMsg := fmt.Sprintf("Failed to write tokens to db with: %s", err.Error())

		comm.AbortBadRequest(ctx, respMsg, comm.CodeInvalidArgs)
		return
	}

	comm.OK(ctx, "Token registered", comm.CodeSuccess)
}
