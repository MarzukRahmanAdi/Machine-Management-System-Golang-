package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Data struct {
	ID           uint `gorm:"primary_key"`
	No1          string
	No2          string
	StartTime    time.Time
	RunningHours float64
	StopTime     time.Time
	Location     string
	Stopped      bool
	StoppedTime  *time.Time
	Remark       *string
}

type ArchivedData struct {
	Date string
	Data []Data
}

func main() {

	r := gin.Default()
	r.Static("/assets", "./assets")

	db, err := gorm.Open("sqlite3", "fac.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	db.AutoMigrate(&Data{})
	// db.Create(&Data{No1: "Value1", No2: "Value2", StartTime: time.Now(), RunningHours: 5, StopTime: time.Now().Add(6 * time.Hour), Location: "Location1"})

	r.LoadHTMLGlob("front-end/*")

	r.GET("/", func(c *gin.Context) {
		var data []Data
		db.Where("stopped = ?", false).Find(&data)
		// r.LoadHTMLFiles("./templates/index.html")
		dataFinal := gin.H{
			"data": data,
		}
		c.HTML(200, "index.teml", dataFinal)
	})

	r.GET("/archieve", func(c *gin.Context) {
		var data []Data
		db.Where("stopped = ?", true).Find(&data)
		archivedData := make(map[string]ArchivedData)

		for _, item := range data {
			date := item.StartTime.Format("2006-01-02")
			if _, exists := archivedData[date]; !exists {
				archivedData[date] = ArchivedData{
					Date: date,
					Data: []Data{item},
				}
			} else {
				existingData := archivedData[date]
				existingData.Data = append(existingData.Data, item)
				archivedData[date] = existingData
			}
		}

		archivedDataSlice := []ArchivedData{}
		for _, v := range archivedData {
			archivedDataSlice = append(archivedDataSlice, v)
		}

		sort.Slice(archivedDataSlice, func(i, j int) bool {
			return archivedDataSlice[i].Date > archivedDataSlice[j].Date
		})

		archieveFinal := gin.H{
			"archivedData": archivedDataSlice,
		}
		c.HTML(200, "archive.teml", archieveFinal)
	})

	r.POST("/add-data", func(c *gin.Context) {
		no1 := c.PostForm("no1")
		no2 := c.PostForm("no2")

		timeInput := c.PostForm("startTime")
		currentTime := time.Now()

		inputTime, err := time.Parse("15:04", timeInput)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(timeInput)

		startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), inputTime.Hour(), inputTime.Minute(), inputTime.Second(), 0, currentTime.Location())

		hoursInput := c.PostForm("hours")
		minutesInput := c.PostForm("minutes")

		hours, _ := strconv.Atoi(hoursInput)

		minutes, _ := strconv.Atoi(minutesInput)

		runningHours := float64(hours) + float64(minutes)/60.0

		location := c.PostForm("location")
		remark := c.PostForm("remark")

		stopTime := startTime.Add(time.Duration(runningHours * float64(time.Hour)))

		var stoppedTime *time.Time = nil

		newData := Data{
			No1:          no1,
			No2:          no2,
			StartTime:    startTime,
			RunningHours: runningHours,
			StopTime:     stopTime,
			Location:     location,
			StoppedTime:  stoppedTime,
			Remark:       &remark,
		}

		db.Create(&newData)

		c.Redirect(302, "/")
	})

	r.POST("/stop-data", func(c *gin.Context) {
		id := c.PostForm("id")

		var data Data
		if err := db.First(&data, id).Error; err != nil {
			c.JSON(400, gin.H{"success": false, "message": "Data not found"})
			return
		}

		if !data.Stopped {
			data.Stopped = true
			stoppedTime := time.Now()
			data.StoppedTime = &stoppedTime
			db.Save(&data)
		}

		c.JSON(200, gin.H{"success": true, "message": "Data stopped"})
	})

	r.Run(":8080")
}
