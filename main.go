package main

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"google.golang.org/api/option"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
)

func emailIsvalid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

const CtxVarMongoDBClient = "mongo-db"
const CtxVarFirebaseApp = "fb-app"
const CtxVarUserId = "user-id"
const CtxVarAuthToken = "auth-token"

func handleCheck(ctx *gin.Context) {
	fmt.Printf("Handle check... \n")

	authHeader := ctx.GetHeader("Authorization")
	authComps := strings.Split(authHeader, " ")
	if len(authComps) != 2 && authComps[0] != "Bearer" {
		fmt.Printf("Invalid header \n")
		ctx.String(400, "Invalid header")
		return
	}

	fbApp := ctx.MustGet(CtxVarFirebaseApp).(*firebase.App)
	fbAuth, _ := fbApp.Auth(ctx)
	_, err := fbAuth.VerifyIDToken(ctx, authComps[1])

	if err != nil {
		fmt.Printf("Unauthorized with %s \n", err.Error())
		ctx.String(401, "Unauthorized")
		return
	}

	ctx.String(200, "Authorized")
}

func createDBUserRecords(ctx *gin.Context, uid string, username string, email string) bool {
	mongoClient := ctx.MustGet(CtxVarMongoDBClient).(*mongo.Client)
	db := mongoClient.Database(MongoDBDatabaseName)
	usersCollection := db.Collection(MongoDBUsersCollection)
	usernamesCollection := db.Collection(MongoDBUsernamesCollection)

	wc := writeconcern.Majority()
	txOptions := options.Transaction().SetWriteConcern(wc)
	session, err := mongoClient.StartSession()
	if err != nil {
		fmt.Printf("Failed to start session \n")
		ctx.String(400, apiResponse("Failed to start tx", CodeCantCreateAuthUser))
		return false
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(
		ctx,
		func(ctx mongo.SessionContext) (interface{}, error) {
			_, err := usersCollection.InsertOne(ctx, UserData{
				Id:       uid,
				Username: username,
				Email:    email,
			})
			if err != nil {
				return nil, err
			}
			_, err = usernamesCollection.InsertOne(ctx, UsernameData{
				Id:     username,
				UserID: uid,
			})

			return nil, err
		},
		txOptions,
	)
	if err != nil {

		fmt.Printf("Tx failed with error: %s\n", err.Error())
		ctx.String(400, apiResponse("Failed to create db records", CodeCantCreateAuthUser))
		return false
	}
	return true
}

type RegisterParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Pwd      string `json:"pwd"`
}

type UserData struct {
	Id       string            `json:"uid" bson:"_id"`
	Username string            `json:"username" bson:"username"`
	Email    string            `json:"email" bson:"email"`
	Img_url  string            `json:"img_url" bson:"img_url"`
	Tokens   map[string]string `json:"tokens" bson:"tokens"`
}
type UsernameData struct {
	Id     string `json:"_id" bson:"_id"`
	UserID string `json:"user_id"`
}

func handleRegister(ctx *gin.Context) {
	//TODO check email conforms requirements

	params := RegisterParams{}

	if ctx.ShouldBind(&params) == nil {
		if !emailIsvalid(params.Email) {
			ctx.String(400, apiResponse("Invalid email format", CodeInvalidArgs))
			return
		}

		fbApp := ctx.MustGet(CtxVarFirebaseApp).(*firebase.App)
		mongoClient := ctx.MustGet(CtxVarMongoDBClient).(*mongo.Client)

		database := mongoClient.Database(MongoDBDatabaseName)
		usernamesCollection := database.Collection(MongoDBUsernamesCollection)

		filter := bson.D{{"_id", params.Username}}
		usernameRes := usernamesCollection.FindOne(ctx, filter)
		if usernameRes.Err() == nil {
			ctx.String(400, apiResponse("User already exists", CodeUsernameTaken))
			return
		}

		if strings.Contains(params.Username, " ") || len(params.Username) < 4 {
			ctx.String(400, apiResponse("Invalid username", CodeUsernameFormatNotValid))
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
			ctx.String(400, apiResponse(fmt.Sprintf("Backend failed to create new user with %s", err.Error()), CodeCantCreateAuthUser))
			return
		}

		if !createDBUserRecords(ctx, userRecord.UID, params.Username, params.Email) {
			return
		}

		ctx.String(200, apiResponse("Registered", CodeSuccess))
	}
}

type UpdateAvatarParams struct {
	ImgUrl string `json:"img_url"`
}

func handleUpdateAvatar(ctx *gin.Context) {
	params := UpdateAvatarParams{}
	err := ctx.ShouldBind(&params)
	if err != nil || len(params.ImgUrl) == 0 {
		ctx.String(400, apiResponse("Invalid args", CodeInvalidArgs))
		return
	}

	userId := ctx.MustGet(CtxVarUserId).(string)
	if len(userId) == 0 {
		return //not authenticated
	}
	mongoClient := ctx.MustGet(CtxVarMongoDBClient).(*mongo.Client)

	filter := bson.D{{"_id", userId}}
	update := bson.D{{"$set", bson.D{{"img_url", params.ImgUrl}}}}
	_, err = mongoClient.Database(MongoDBDatabaseName).Collection(MongoDBUsersCollection).UpdateOne(ctx, filter, update)

	if err != nil {
		ctx.String(400, apiResponse("Failed to update url", CodeInvalidArgs))
		return
	}
	ctx.String(200, apiResponse("Success", CodeSuccess))
}

type CompleteRegisterParams struct {
	Username string `json:"username"`
}

func handleCompleteRegister(ctx *gin.Context) {
	params := CompleteRegisterParams{}
	if ctx.ShouldBind(&params) != nil {
		ctx.String(400, apiResponse("Invalid args", CodeInvalidArgs))
		return
	}
	userId := ctx.MustGet(CtxVarUserId).(string)
	if len(userId) == 0 {
		return // not authenticated
	}
	authToken := ctx.MustGet(CtxVarAuthToken).(*auth.Token)
	fbApp := ctx.MustGet(CtxVarFirebaseApp).(*firebase.App)
	auth, _ := fbApp.Auth(ctx)
	userRecord, err := auth.GetUser(ctx, authToken.UID)
	fmt.Printf(userRecord.ProviderID + "\n")
	fmt.Printf(authToken.Firebase.SignInProvider + "\n")
	if err != nil {
		ctx.String(401, apiResponse("Smth went wrong", CodeNotAuthenticated))
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
		ctx.String(401, apiResponse("Invalid provider", CodeNotAuthenticated))
		return
	}

	if !createDBUserRecords(ctx, authToken.UID, params.Username, email) {
		return
	}

	ctx.String(200, apiResponse("Completed registration", CodeSuccess))
}

type RegisterTokenParams struct {
	Token      string `json:"token"`
	DeviceName string `json:"device_name"`
}

func handleRegisterToken(ctx *gin.Context) {
	userId := ctx.MustGet(CtxVarUserId).(string)
	if len(userId) == 0 {
		return // not authenticated
	}
	params := RegisterTokenParams{}
	err := ctx.ShouldBind(&params)
	if err != nil {
		fmt.Printf("Couldnt bind register token params: %s\n", err.Error())
		ctx.String(400, apiResponse("Invalid args", CodeInvalidArgs))
		return
	}

	mongoClient := ctx.MustGet(CtxVarMongoDBClient).(*mongo.Client)
	usersCollection := mongoClient.Database(MongoDBDatabaseName).Collection(MongoDBUsersCollection)

	filter := bson.D{{"_id", userId}}
	res := usersCollection.FindOne(ctx, filter)
	if res.Err() != nil {
		ctx.String(400, apiResponse("Auth error", CodeNotAuthenticated))
		return
	}
	userData := UserData{}
	err = res.Decode(&userData)
	if err != nil {
		ctx.String(400, apiResponse("Failed to decode data from db", CodeInvalidArgs))
		return
	}
	if userData.Tokens == nil {
		userData.Tokens = make(map[string]string)
	}
	userData.Tokens[params.DeviceName] = params.Token

	update := bson.D{{"$set", bson.D{{"tokens", userData.Tokens}}}} //bson.D{$set: {"tokens", userData.Tokens}}
	_, err = usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		ctx.String(400, apiResponse(fmt.Sprintf("Failed to write tokens to db with: %s", err.Error()), CodeInvalidArgs))
		return
	}

	ctx.String(200, apiResponse("Token registered", CodeSuccess))
}

type AddMessageParams struct {
	Msg  string `json:"text"`
	ToId string `json:"to"`
}

type Ids []string

type MessageData struct {
	ToId           string `bson:"to" json:"to"`
	FromId         string `bson:"from" json:"from"`
	Text           string `bson:"msg" json:"msg"`
	FromUsername   string `bson:"username" json:"username"`
	ImgUrl         string `bson:"img_url" json:"img_url"`
	CreatedAt      int64  `bson:"created_at" json:"created_at"`
	ConversationID string `bson:"conv_id" json:"conv_id"`
}

func getUserData(ctx *gin.Context, id string, data *UserData) bool {
	mongoClient := ctx.MustGet(CtxVarMongoDBClient).(*mongo.Client)
	db := mongoClient.Database(MongoDBDatabaseName)
	usersCollection := db.Collection(MongoDBUsersCollection)

	filter := bson.D{{"_id", id}}

	res := usersCollection.FindOne(ctx, filter)
	if res.Err() != nil {
		ctx.String(400, apiResponse("Invalid user or user is not correctly registered", CodeNotAuthenticated))
		return false
	}
	if data != nil {
		err := res.Decode(data)
		if err != nil {
			ctx.String(400, apiResponse("Cant decode data from db", CodeInvalidArgs))
			return false
		}
	}

	return true
}

func composeChatKey(uid1 string, uid2 string) string {
	ids := []string{uid1, uid2}
	sort.Strings(ids)

	compIndex := ids[0] + ids[1]
	return compIndex
}

func fcmSendNewMessage(ctx *gin.Context, tokens map[string]string, msg MessageData, needsNotification bool, needsMsg bool) bool {

	fbApp := ctx.MustGet(CtxVarFirebaseApp).(*firebase.App)
	fcmClient, _ := fbApp.Messaging(ctx)

	isNotification := 0
	if needsNotification {
		isNotification = 1
	}
	isMsg := 0
	if needsMsg {
		isMsg = 1
	}

	fcmMsgData := map[string]string{
		"is_notification": strconv.FormatInt(int64(isNotification), 10),
		"is_msg":          strconv.FormatInt(int64(isMsg), 10),
	}

	if needsNotification {
		fcmMsgData["click_action"] = "FLUTTER_NOTIFICATION_CLICK"
	}
	if needsMsg {
		fcmMsgData["_from"] = msg.FromId
		fcmMsgData["to"] = msg.ToId
		fcmMsgData["username"] = msg.FromUsername
		fcmMsgData["msg"] = msg.Text
		fcmMsgData["img_url"] = msg.ImgUrl
		fcmMsgData["created_at"] = strconv.FormatInt(msg.CreatedAt, 10)
	}

	for _, token := range tokens {
		fcmMsg := &messaging.Message{}

		fcmMsg.Token = token
		fcmMsg.Data = fcmMsgData
		if needsNotification {
			fcmMsg.Notification = &messaging.Notification{
				Title: msg.FromUsername,
				Body:  msg.Text,
			}
			fcmMsg.Android = &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					Sound: "default",
				},
			}
		}

		_, err := fcmClient.Send(
			ctx,
			fcmMsg,
		)

		if err != nil {
			fmt.Printf("Failed to send FCM message %s \n", err.Error())
			ctx.String(400, apiResponse("Failed to send FCM messages", CodeInvalidArgs))
			return false
		}

	}
	return true
}
func handleAddMessage(ctx *gin.Context) {
	userId := ctx.MustGet(CtxVarUserId).(string)
	if len(userId) == 0 {
		return
	}
	params := AddMessageParams{}
	if ctx.ShouldBind(&params) != nil {
		ctx.String(400, apiResponse("Invalid args", CodeInvalidArgs))
		return
	}

	if len(params.Msg) == 0 {
		ctx.String(400, apiResponse("Message cant be empty", CodeInvalidArgs))
		return
	}

	mongoClient := ctx.MustGet(CtxVarMongoDBClient).(*mongo.Client)
	db := mongoClient.Database(MongoDBDatabaseName)

	fromUserData := UserData{}
	if !getUserData(ctx, userId, &fromUserData) {
		return
	}
	toUserData := UserData{}
	if !getUserData(ctx, params.ToId, &toUserData) {
		return
	}

	compIndex := composeChatKey(userId, params.ToId)

	msg := MessageData{
		ConversationID: compIndex,
		Text:           params.Msg,
		FromId:         userId,
		ToId:           params.ToId,
		FromUsername:   fromUserData.Username,
		ImgUrl:         fromUserData.Img_url,
		CreatedAt:      time.Now().Unix(),
	}

	messagesCollection := db.Collection(MongoDBMessagesCollection)

	_, err := messagesCollection.InsertOne(ctx, msg)

	if err != nil {
		ctx.String(400, apiResponse(fmt.Sprintf("Failed to write messages to db with: %s", err.Error()), CodeInvalidArgs))
		return
	}

	if toUserData.Tokens != nil {
		if !fcmSendNewMessage(ctx, toUserData.Tokens, msg, true, false) {
			return
		}
		if !fcmSendNewMessage(ctx, toUserData.Tokens, msg, false, true) {
			return
		}
	}
	if fromUserData.Tokens != nil {
		if !fcmSendNewMessage(ctx, fromUserData.Tokens, msg, false, true) {
			return
		}
	}

	ctx.String(200, apiResponse("Success", CodeSuccess))
}
func handleGetUser(ctx *gin.Context) {
	userId := ctx.MustGet(CtxVarUserId).(string)
	if len(userId) == 0 {
		return
	}
	targetUserId := ctx.Param("uid")

	userData := UserData{}
	if !getUserData(ctx, targetUserId, &userData) {
		return
	}

	ctx.String(200, apiResponseWithJson("Success", CodeSuccess, userData))
}
func handleGetUid(ctx *gin.Context) {
	userId := ctx.MustGet(CtxVarUserId).(string)
	if len(userId) == 0 {
		return
	}

	targetUsername := ctx.Param("username")

	mongoClient := ctx.MustGet(CtxVarMongoDBClient).(*mongo.Client)
	db := mongoClient.Database(MongoDBDatabaseName)
	usernamesCollection := db.Collection(MongoDBUsernamesCollection)

	usernameData := UsernameData{}

	filter := bson.D{{"_id", targetUsername}}
	res := usernamesCollection.FindOne(ctx, filter)

	if res.Err() != nil {
		ctx.String(400, apiResponse("No such user", CodeUserNotRegistered))
		return
	}

	err := res.Decode(&usernameData)
	if err != nil {
		ctx.String(400, apiResponse("Failed to decode server response", CodeUserNotRegistered))
		return
	}

	ctx.String(200, apiResponseWithJson("Success", CodeSuccess, usernameData))
}
func InjectParams(app *firebase.App, mongoClient *mongo.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(CtxVarFirebaseApp, app)
		ctx.Set(CtxVarMongoDBClient, mongoClient)
	}
}
func AuthMiddleware(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	authComps := strings.Split(authHeader, " ")
	if len(authComps) != 2 && authComps[0] != "Bearer" {
		fmt.Printf("Invalid header \n")

		ctx.String(400, apiResponse("Invalid header", CodeNotAuthenticated)) //"Invalid header")
		return
	}

	fbApp := ctx.MustGet(CtxVarFirebaseApp).(*firebase.App)
	fbAuth, _ := fbApp.Auth(ctx)
	authToken, err := fbAuth.VerifyIDToken(ctx, authComps[1])

	if err != nil {
		fmt.Printf("Unauthorized with: %s \n", err.Error())
		ctx.String(401, apiResponse("Invalid creds", CodeNotAuthenticated))

		ctx.Set(CtxVarUserId, "")
		return
	}

	ctx.Set(CtxVarUserId, authToken.UID) //TODO: remove
	ctx.Set(CtxVarAuthToken, authToken)
}

type GetChatParams struct {
	Limit           int    `json:"limit"`
	BeforeTimeStamp int64  `json:"before_timestamp"`
	With            string `json:"with"`
	Inverse         bool   `json:"inverse"`
}

func handleGetChat(ctx *gin.Context) {
	userId := ctx.MustGet(CtxVarUserId).(string)
	if len(userId) == 0 {
		return
	}

	params := GetChatParams{}
	if ctx.ShouldBind(&params) != nil {
		fmt.Printf("Invalid params\n")
		ctx.String(400, apiResponse("Invalid args", CodeInvalidArgs))
		return
	}

	fmt.Printf("Getting msgs before %d\n", params.BeforeTimeStamp)

	mongoClient := ctx.MustGet(CtxVarMongoDBClient).(*mongo.Client)
	db := mongoClient.Database(MongoDBDatabaseName)
	messagesCollection := db.Collection(MongoDBMessagesCollection)

	// TODO dont allow high limit to mitigate possible attack
	opts := options.Find().SetLimit(int64(params.Limit)).SetSort(bson.D{{"created_at", -1}})

	compKey := composeChatKey(userId, params.With)

	fmt.Printf("Gettting chat with comp key %s\n", compKey)

	filter := bson.D{
		{
			"$and", bson.A{
				bson.D{{"conv_id", compKey}},
				bson.D{{"created_at", bson.D{{"$lte", params.BeforeTimeStamp}}}},
			},
		}}
	cursor, err := messagesCollection.Find(ctx, filter, opts)
	if err != nil {
		ctx.String(400, apiResponse("Failed to fetch messages", CodeInvalidArgs))
		return
	}

	var messages []MessageData

	err = cursor.All(ctx, &messages)
	if err != nil {
		ctx.String(400, apiResponse("Couldnt parse data from db", CodeInvalidArgs)) //TODO add code
		return
	}

	ctx.String(200, apiResponseWithJson("Success", CodeSuccess, messages))
}

func main() {
	godotenv.Load()

	credsFileName, ok := os.LookupEnv("FIREBASE_CREDS_FILE")
	if !ok {
		log.Fatal("Service account is required to be set through env var file path to creds file")
	}
	mongodbConnectUrl, ok := os.LookupEnv("MONGODB_CONNECT_URL")
	if !ok {
		log.Fatal("Mongodb connection url with creds should be set thorugh env file")
	}

	opt := option.WithCredentialsFile(credsFileName)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Failed to create Firebase app, with %s", err.Error())
	}

	mongoServerAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongodbConnectUrl).SetServerAPIOptions(mongoServerAPI)
	mongoClient, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatal("Failed to connect to mongo db")
	}

	g := gin.Default()
	g.Use(InjectParams(app, mongoClient))
	g.POST("/register", handleRegister)
	g.POST("/completeregister", AuthMiddleware, handleCompleteRegister)
	g.POST("/registertoken", AuthMiddleware, handleRegisterToken)
	g.POST("/updateavatar", AuthMiddleware, handleUpdateAvatar)
	g.POST("/addmessage", AuthMiddleware, handleAddMessage)
	g.POST("/check", handleCheck)
	g.POST("/chat", AuthMiddleware, handleGetChat) //IDIOTIC DART DOESN'T ENABLE GET REQUESTS WITH BODY
	g.GET("/users/id/:uid", AuthMiddleware, handleGetUser)
	g.GET("/uid/:username", AuthMiddleware, handleGetUid)
	g.Run(":8080")
}
