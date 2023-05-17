package controllers

import (
	"FaceRecognition/models"
	"encoding/json"
	"fmt"
	"github.com/Kagami/go-face"
	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
	"sync"
)

type FaceController interface {
	GetAll(ctx *gin.Context)

	Update(ctx *gin.Context)

	Create(ctx *gin.Context)

	Delete(ctx *gin.Context)

	DeleteAll(ctx *gin.Context)

	UpLoad(ctx *gin.Context)

	Recognize(ctx *gin.Context)
}

type controller struct {
	Faces       []models.Face
	RecEngine   *face.Recognizer
	FaceMap     map[int32]string
	FaceDesLibs []face.Descriptor
	FaceID      []int32
	mtx         sync.Mutex
	userName string
	password string
}

var imagesDir = filepath.Join("resources", "images")
var jsonFileName = filepath.Join(imagesDir, "faceinfo.json")

type generator struct {
	counter int
	mtx     sync.Mutex
}

func (g *generator) generatorNextID() int {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	g.counter++
	return g.counter
}

var g *generator = &generator{}

func NewFaceController(faceLib []models.Face, recognizer *face.Recognizer, faceMap map[int32]string, faceDesLib []face.Descriptor, faceid []int32) FaceController {
	return &controller{Faces: faceLib, RecEngine: recognizer, FaceMap: faceMap, FaceDesLibs: faceDesLib, FaceID: faceid,userName : "ppd",password:"123456"}
}

func (c *controller) GetAll(ctx *gin.Context) {
	//TODO implement me
	ctx.JSON(200, c.FaceMap)
}

func (c *controller) Update(ctx *gin.Context) {
	//TODO implement me
	var faceToUpdate models.Face
	if err := ctx.ShouldBindUri(&faceToUpdate); err != nil {
		ctx.String(400, "bad request %v", err)
		return
	}
	if err := ctx.BindJSON(&faceToUpdate); err != nil {
		ctx.String(400, "bad json %v", err)
		return
	}
	for index, face := range c.Faces {
		if face.Id == faceToUpdate.Id {
			c.Faces[index] = faceToUpdate
			ctx.String(200, "success to update")
			return
		}
	}
	ctx.String(400, "bad request cannot find face with id is %d to update", faceToUpdate.Id)
}

func (c *controller) Create(ctx *gin.Context) {
	//TODO implement me
	var newFace = models.Face{Id: g.generatorNextID()}
	if err := ctx.BindJSON(&newFace); err != nil {
		ctx.String(400, "bad request %v", err)
		return
	}
	c.Faces = append(c.Faces, newFace)
	ctx.String(200, "success,create new face,new face is %d", newFace.Id)
}

func (c *controller) Delete(ctx *gin.Context) {
	//TODO implement me
	var faceToDelete models.Face
	if err := ctx.BindUri(&faceToDelete); err != nil {
		ctx.String(400, "bad request", err)
		return
	}
	for index, face := range c.Faces {
		if face.Id == faceToDelete.Id {
			c.Faces = append(c.Faces[:index], c.Faces[index+1:len(c.Faces)]...)
			ctx.String(200, "success,delete id is %d face", faceToDelete.Id)
			return
		}
	}
}

func (c *controller) DeleteAll(ctx *gin.Context) {
	m := make(map[string]interface{},0)
	if err := ctx.BindJSON(&m); err != nil {
		ctx.String(400,"bad request",err)
		return
	}
	if val,ok := m["userName"];ok{
		if val != c.userName {
			ctx.String(400,"bad request,username is not exist")
			return
		}
	}else {
		ctx.String(400,"bad request,no username")
		return
	}
	if val,ok :=m["password"];ok {
		if val != c.password {
			ctx.String(400,"bad request,password error")
			return
		}
	}else {
		ctx.String(400,"bad request,no password")
		return
	}
	//清空所有数据
	c.mtx.Lock();
	defer c.mtx.Unlock();
	c.Faces = c.Faces[0:0];
	c.FaceID = c.FaceID[0:0];
	c.FaceDesLibs = c.FaceDesLibs[0:0];
	for k := range c.FaceMap {
		delete(c.FaceMap,k)
	}
	err := os.Truncate(jsonFileName, 0)
	if err != nil {
		fmt.Println("clear file error,",err)
		ctx.String(200,"delete file data error + %v",err)
		return
	}
	ctx.String(200,"sucess delete all faces data")
	
}

func (c *controller) UpLoad(ctx *gin.Context) {
	filename := ctx.PostForm("fileName")
//	id := ctx.PostForm("id")
//	num, _ := strconv.Atoi(id)
	name := ctx.PostForm("name")
    fmt.Println("people name:",name)
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.String(400, "file upload failed,please upload again:%v", err)
		return
	}
	ctx.String(200, "upload success %v", file.Filename)
	newFilePath := filepath.Join(imagesDir, file.Filename)
	err = ctx.SaveUploadedFile(file, newFilePath)
	if err != nil {
		fmt.Println("save net file error", err)
		return
	}
    fmt.Println("save file success,file name is:",newFilePath)
	go func() {
		faceInfos, err := c.RecEngine.RecognizeFile(newFilePath)
		if err != nil {
			fmt.Println("Recognize ", file.Filename, " error", err)
			return
		}
        fmt.Println("Recognize success");
		newId := 0
		for {
			newId = g.generatorNextID()
			if _, ok := c.FaceMap[int32(newId)]; ok {
				continue
			} else {
				break
			}
		}
		

		for _, faceInfo := range faceInfos {
			newface := &models.Face{
				Id:         newId,
				Name:       name,
				File:       filename,
				Shapes:     nil,
				Descriptor: faceInfo.Descriptor,
			}
			c.mtx.Lock()

			c.FaceDesLibs = append(c.FaceDesLibs, faceInfo.Descriptor)
			c.FaceID = append(c.FaceID, int32(newId))
			//更新人脸库
			c.RecEngine.SetSamples(c.FaceDesLibs, c.FaceID)
			//更新人脸结构数组
			c.Faces = append(c.Faces, *newface)
			//更新json文件
			c.FaceMap[int32(newId)] = name

			buf, err := json.Marshal(c.Faces)
			if err != nil {
				fmt.Println("json marshal error", err)
				c.mtx.Unlock()
				return
			}
			 err = os.WriteFile(jsonFileName,buf,0666)
			if err != nil {
				fmt.Println("write to ", jsonFileName, " error", err)
				c.mtx.Unlock()
				return
			}
            fmt.Println("write file success")
			c.mtx.Unlock()

		}
	}()

}

func (c *controller) Recognize(ctx *gin.Context) {
	var faceToRecognize models.Face
	if err := ctx.BindJSON(&faceToRecognize); err != nil {
		ctx.String(400, "bad request,%v", err)
		return
	}
	id := c.RecEngine.Classify(faceToRecognize.Descriptor)
	if id < 0 {
		ctx.JSON(200, gin.H{
			"name": "unknown",
		})
	} else {
		if val, ok := c.FaceMap[int32(id)]; ok {
			ctx.JSON(200, gin.H{
				"name": val,
			})
			return
		}
		ctx.JSON(200, gin.H{
			"name": "unknown",
		})
		return
	}

}
