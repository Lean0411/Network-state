package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"sync"
)

var (
	profileStore = make(map[string]*Profile)
	profileMutex sync.Mutex
	profileID    int
)

type Profile struct {
	ID             string         `json:"profile_id"`
	Name           string         `json:"name"`
	BestQuality    *QualityConfig `json:"bestQuality,omitempty"`
	SatisfyQuality *QualityConfig `json:"satisfyQuality,omitempty"`
}

type QualityConfig struct {
	PingConfiguration *PingConfig  `json:"pingConfiguration,omitempty"`
	Indicator         string       `json:"indicator,omitempty"`
	SLAProfile        *SLAProfile  `json:"slaProfile,omitempty"`
}

type PingConfig struct {
	Period     int    `json:"period,omitempty"`
	PingTarget string `json:"pingTarget,omitempty"`
}

type SLAProfile struct {
	BandwidthOptimized *ThresholdWrapper `json:"bandwidthOptimized,omitempty"`
	LatencySensitive   *ThresholdWrapper `json:"latencySensitive,omitempty"`
	QualityStability   *ThresholdWrapper `json:"qualityStability,omitempty"`
	ReliabilityMonitor *ThresholdWrapper `json:"reliabilityMonitor,omitempty"`
}

type ThresholdWrapper struct {
	Threshold *Threshold `json:"threshold,omitempty"`
}

type Threshold struct {
	Upload         int     `json:"upload,omitempty"`
	Download       int     `json:"download,omitempty"`
	Delay          int     `json:"delay,omitempty"`
	Jitter         float64 `json:"jitter,omitempty"`
	PacketLossRate int     `json:"packetLossRate,omitempty"`
}

type ProfileListItem struct {
    ProfileID  string `json:"profile_id"`
    Name       string `json:"name"`
    Mode       string `json:"mode"`
    PingTarget string `json:"ping_target"`
    Period     int    `json:"period"`
    Indicator  string `json:"indicator"`
}

// 從請求中解析配置資料，然後在profileStore中新增配置。
func addProfile(c *gin.Context) {
	var profile Profile
	if err := c.BindJSON(&profile); err != nil {
		c.JSON(400, gin.H{"status": false, "err_message": err.Error()})
		return
	}

	profileMutex.Lock()
	defer profileMutex.Unlock()

	profileID++
	profile.ID = strconv.Itoa(profileID)
	profileStore[profile.ID] = &profile

	c.JSON(200, gin.H{"status": true})
}

// 解析 profile_id，然後在profileStore 中找到並回傳該配置。
func getProfile(c *gin.Context) {
	var request struct {
		ProfileID string `json:"profile_id"`
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"status": false, "err_message": err.Error()})
		return
	}

	profileMutex.Lock()
	profile, exists := profileStore[request.ProfileID]
	profileMutex.Unlock()

	if !exists {
		c.JSON(404, gin.H{"status": false, "err_message": "Profile not found"})
		return
	}

	c.JSON(200, profile)
}

//解析 profile_id和其他編輯資料，然後在profileStore中更新配置。
func editProfile(c *gin.Context) {
    var request struct {
        ProfileID    string         `json:"profileId"`
        Name         string         `json:"name"`
        BestQuality  *QualityConfig `json:"bestQuality,omitempty"`
        SatisfyQuality *QualityConfig `json:"satisfyQuality,omitempty"`
    }
    if err := c.BindJSON(&request); err != nil {
        c.JSON(400, gin.H{"status": false, "err_message": err.Error()})
        return
    }

    profileMutex.Lock()
    defer profileMutex.Unlock()

    profile, exists := profileStore[request.ProfileID]
    if !exists {
        c.JSON(404, gin.H{"status": false, "err_message": "Profile not found"})
        return
    }

    if request.Name != "" {
        profile.Name = request.Name
    }

    if request.BestQuality != nil {
        profile.BestQuality = request.BestQuality
    }

    if request.SatisfyQuality != nil {
        profile.SatisfyQuality = request.SatisfyQuality
    }

    c.JSON(200, gin.H{"status": true})
}

//刪除
func deleteProfile(c *gin.Context) {
    var request struct {
        ProfileID string `json:"profile_id"`
    }
    if err := c.BindJSON(&request); err != nil {
        c.JSON(400, gin.H{"status": false, "err_message": err.Error()})
        return
    }

    profileMutex.Lock()
    defer profileMutex.Unlock()

    _, exists := profileStore[request.ProfileID]
    if !exists {
        c.JSON(404, gin.H{"status": false, "err_message": "Profile not found"})
        return
    }

    delete(profileStore, request.ProfileID)
    c.JSON(200, gin.H{"status": true})
}

//列出ProfileListItem結構
func listProfiles(c *gin.Context) {
    profileMutex.Lock()
    defer profileMutex.Unlock()

    profileList := make([]ProfileListItem, 0, len(profileStore))

    for _, profile := range profileStore {
        var mode string
        var pingConfig *PingConfig
        var indicator string

        if profile.BestQuality != nil {
            mode = "best_quality"
            pingConfig = profile.BestQuality.PingConfiguration
            indicator = profile.BestQuality.Indicator
        } else if profile.SatisfyQuality != nil {
            mode = "satisfy_quality"
            pingConfig = profile.SatisfyQuality.PingConfiguration
            indicator = profile.SatisfyQuality.Indicator
        }

        if pingConfig != nil {
            item := ProfileListItem{
                ProfileID:  profile.ID,
                Name:       profile.Name,
                Mode:       mode,
                PingTarget: pingConfig.PingTarget,
                Period:     pingConfig.Period,
                Indicator:  indicator,
            }
            profileList = append(profileList, item)
        }
    }

    c.JSON(200, gin.H{
        "status": true,
        "data":   profileList,
    })
}


func main() {
	r := gin.Default()

	r.POST("/api/v1/wanBinding/slaprofile/add", addProfile)
	r.POST("/api/v1/wanBinding/slaprofile/content", getProfile)
	r.POST("/api/v1/wanBinding/slaprofile/edit", editProfile)
	r.POST("/api/v1/wanBinding/slaprofile/delete", deleteProfile)
	r.GET("/api/v1/wanBinding/slaprofile/list", listProfiles)

	r.Run(":8080")
}
