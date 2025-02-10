// package for user files upload functionality
package upload

import (
	"bytes"
	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"io"
	"net/http"

	rawstorage "cloud.google.com/go/storage"
	"firebase.google.com/go/v4/storage"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func RegisterHandlers(authRoutes *gin.RouterGroup, publicRoutes *gin.RouterGroup) {
	authRoutes.POST("/updateavatarfile", accounts.CompleteRegisteredMiddleware, handleUpdateAvatarFile) // TODO: rename path to /avatar
}

func handleUpdateAvatarFile(ctx *gin.Context) {
	userId := ctx.MustGet(auth.CtxVarUserId).(string)
	fbStorage := ctx.MustGet(accounts.CtxVarFirebaseStorage).(*storage.Client)
	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	const imageMaxMBs = 5

	reader := io.LimitReader(ctx.Request.Body, imageMaxMBs*1000000)
	data, err := io.ReadAll(reader)
	if err != nil {
		errMsg := "Failed to load file"
		if err.Error() == "EOF" {
			errMsg = "Maximum allowed image size is 5 MBs" // could do interpolation with imageMaxMBs, but dont want extra allocations
		}
		comm.AbortBadRequest(ctx, errMsg, comm.CodeInvalidArgs)
		return
	}

	mimeType := http.DetectContentType(data)
	if len(mimeType) < 5 || mimeType[0:5] != "image" {
		comm.AbortBadRequest(ctx, "Wrong file type", comm.CodeInvalidArgs)
		return
	}
	choseExt := false
	extMimePart := mimeType[6:]
	ext := ".jpg"
	if extMimePart == "jpeg" {
		choseExt = true
	}
	if extMimePart == "png" {
		choseExt = true
		ext = ".png"
	}
	if !choseExt {
		comm.AbortBadRequest(ctx, "Image type not supported", comm.CodeInvalidArgs)
		return
	}

	bucket, err := fbStorage.DefaultBucket()
	if err != nil {
		comm.AbortBadRequest(ctx, "Failed to access storage bucket. "+err.Error(), comm.CodeInvalidArgs)
		return
	}
	obj := bucket.Object("user_images/" + userId + ext)
	w := obj.NewWriter(ctx)
	w.ObjectAttrs.CacheControl = "public,max-age=86400" // 86400
	w.ObjectAttrs.ContentType = mimeType

	_, err = io.Copy(w, bytes.NewBuffer(data))
	if err != nil {
		comm.AbortBadRequest(ctx, "Failed to upload data to storage. "+err.Error(), comm.CodeInvalidArgs)
		return
	}
	err = w.Close()
	if err != nil {
		comm.AbortBadRequest(ctx, "Failed to upload data to storage on close. "+err.Error(), comm.CodeInvalidArgs)
		return
	}
	if err := obj.ACL().Set(ctx, rawstorage.AllUsers, rawstorage.RoleReader); err != nil {
		comm.AbortBadRequest(ctx, "Failed to set ACL. "+err.Error(), comm.CodeInvalidArgs)
		return
	}
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		comm.AbortBadRequest(ctx, "Failed to retrive attributes. "+err.Error(), comm.CodeInvalidArgs)
		return
	}

	filter := bson.D{{Key: "_id", Value: userId}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "img_url", Value: attrs.MediaLink}}}}
	_, err = mongoInst.Collection(database.UsersCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		comm.AbortBadRequest(ctx, "Failed to update url", comm.CodeInvalidArgs)
		return
	}

	comm.GenericOK(ctx)
}
