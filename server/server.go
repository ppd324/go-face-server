package main

import (
	"FaceRecognition/controllers"
	"FaceRecognition/models"
	"encoding/json"
	"fmt"
	"github.com/Kagami/go-face"
	"github.com/gin-gonic/gin"
	"os"

	//"github.com/ppd324/go-face"
	"log"
	"path/filepath"
)

const dataDir = "resources"

var (
	modelsDir = filepath.Join(dataDir, "models")
	imagesDir = filepath.Join(dataDir, "images")
	faceInfoFile = filepath.Join(imagesDir, "faceinfo.json")
)

func loadFaceInfoFile() ([]models.Face, []face.Descriptor, []int32, map[int32]string) {
	faceLib := make([]models.Face, 0)
	facedeslib := make([]face.Descriptor, 0)
	faceId := make([]int32, 0)
	faceMap := make(map[int32]string, 0)
	buf, err := os.ReadFile(faceInfoFile)
	if err != nil || buf == nil {
		fmt.Println("load file failed,err:", err)
		return faceLib, facedeslib, faceId, faceMap
    }
	err = json.Unmarshal(buf, &faceLib)
	if err != nil {
		fmt.Println("json unmarshal failed,err:", err)
		return faceLib, facedeslib, faceId, faceMap
	}
	for _, fac := range faceLib {
		facedeslib = append(facedeslib, fac.Descriptor)
		fmt.Println(fac.Id)
		faceId = append(faceId, int32(fac.Id))
		faceMap[int32(fac.Id)] = fac.Name
	}
	return faceLib, facedeslib, faceId, faceMap

}
func main() {
	//gin.SetMode(gin.ReleaseMode)
	facelib, facesDescribArr, faceId, facesMap := loadFaceInfoFile()
	server := gin.Default()
	rec, err := face.NewRecognizer(modelsDir)
	if err != nil {
		log.Fatalf("Can't init fa recognizer: %v", err)
	}
	fmt.Println("face engine start success")
	Image := filepath.Join(imagesDir, "zhou.jpg")
	fmt.Println("Image path is: ", Image)
	faces, err := rec.RecognizeFile(Image)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("has %d faces\n", len(faces))
	for _, fa := range faces {
		for index, pointer := range fa.Shapes {
			fmt.Printf("%d shape [%d:%d] ", index, pointer.X, pointer.Y)
		}
		newFace := models.NewFace(0, "周杰伦", "zhou.jpg", nil, fa.Descriptor)
		facelib = append(facelib, *newFace)
        facesDescribArr = append(facesDescribArr,fa.Descriptor)
        faceId = append(faceId,0)
        facesMap[0] = "周杰伦"
		//fmt.Println()
	}

	buf, err := json.Marshal(facelib)
	if err != nil {
		fmt.Println("json marshal error", err)
	}
	err = os.WriteFile(faceInfoFile, buf, 0666)
	if err != nil {
		fmt.Println("write to file error", err)
	}

	//// Free the resources when you're finished.

	defer rec.Close()
	FaceController := controllers.NewFaceController(facelib, rec, facesMap, facesDescribArr, faceId)
	FaceGroup := server.Group("/faces")
	FaceGroup.GET("/", FaceController.GetAll)
	FaceGroup.PUT("/:id", FaceController.Update)
	FaceGroup.DELETE("/:id", FaceController.Delete)
	FaceGroup.POST("/", FaceController.Create)
	FaceGroup.POST("/uploadImage/", FaceController.UpLoad)
	FaceGroup.POST("/recognition", FaceController.Recognize)
	FaceGroup.DELETE("/all",FaceController.DeleteAll)
	fmt.Println("i am server")
	server.Run(":8000")
}
