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
	var profile struct {
		Name         string         `json:"name"`
		Mode         string         `json:"mode"` // best_quality, satisfy_quality
		BestQuality  *QualityConfig `json:"bestQuality,omitempty"`
		SatisfyQuality *QualityConfig `json:"satisfyQuality,omitempty"`
	}
	if err := c.BindJSON(&profile); err != nil {
		c.JSON(400, gin.H{"status": false, "err_message": err.Error()})
		return
	}

	// 模式驗證：best_quality
	if profile.Mode == "best_quality" {
		if profile.BestQuality == nil ||
			profile.BestQuality.PingConfiguration == nil ||
			profile.BestQuality.Indicator == "" {
			c.JSON(400, gin.H{
				"status": false,
				"err_message": "When mode is best_quality, bestQuality.pingConfiguration and bestQuality.indicator must be provided",
			})
			return
		}
		if profile.BestQuality.SLAProfile != nil {
			c.JSON(400, gin.H{
				"status": false,
				"err_message": "When mode is best_quality, bestQuality.slaProfile should not be provided",
			})
			return
		}
	}

	// 模式驗證：satisfy_quality
	if profile.Mode == "satisfy_quality" {
		if profile.SatisfyQuality == nil ||
			profile.SatisfyQuality.PingConfiguration == nil ||
			profile.SatisfyQuality.SLAProfile == nil {
			c.JSON(400, gin.H{
				"status": false,
				"err_message": "When mode is satisfy_quality, satisfyQuality.pingConfiguration and satisfyQuality.slaProfile must be provided",
			})
			return
		}
		if profile.SatisfyQuality.SLAProfile.BandwidthOptimized == nil &&
			profile.SatisfyQuality.SLAProfile.LatencySensitive == nil &&
			profile.SatisfyQuality.SLAProfile.QualityStability == nil &&
			profile.SatisfyQuality.SLAProfile.ReliabilityMonitor == nil {
			c.JSON(400, gin.H{
				"status": false,
				"err_message": "When mode is satisfy_quality, at least one satisfyQuality.slaProfile item must be provided",
			})
			return
		}
	}

	profileMutex.Lock()
	defer profileMutex.Unlock()

	profileID++
	profileIDStr := strconv.Itoa(profileID)
	// 使用適當的 Profile 物件儲存在 profileStore 中
	profileStore[profileIDStr] = &Profile{
		ID:             profileIDStr,
		Name:           profile.Name,
		BestQuality:    profile.BestQuality,
		SatisfyQuality: profile.SatisfyQuality,
	}

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
    // 嘗試將接收到的JSON綁定到結構體並檢查錯誤
    if err := c.BindJSON(&request); err != nil {
        c.JSON(400, gin.H{"status": false, "err_message": err.Error()})
        return
    }

    // 驗證 bestQuality 字段
    if request.BestQuality != nil {
        // 確保 pingConfiguration 和 indicator 被設置
        if request.BestQuality.PingConfiguration == nil ||
            request.BestQuality.Indicator == "" {
            c.JSON(400, gin.H{
                "status": false,
                "err_message": "當提供bestQuality時,必須設置bestQuality.pingConfiguration和bestQuality.indicator",
            })
            return
        }
    }

    // 驗證 satisfyQuality 字段
    if request.SatisfyQuality != nil {
        // 確保 pingConfiguration 和 slaProfile 被設置
        if request.SatisfyQuality.PingConfiguration == nil ||
            request.SatisfyQuality.SLAProfile == nil {
            c.JSON(400, gin.H{
                "status": false,
                "err_message": "當提供satisfyQuality時,必須設置satisfyQuality.pingConfiguration和satisfyQuality.slaProfile",
            })
            return
        }

        // 確保至少有一個 slaProfile 子項目被提供並設置了閾值
        if request.SatisfyQuality.SLAProfile.BandwidthOptimized == nil &&
            request.SatisfyQuality.SLAProfile.LatencySensitive == nil &&
            request.SatisfyQuality.SLAProfile.QualityStability == nil &&
            request.SatisfyQuality.SLAProfile.ReliabilityMonitor == nil {
            c.JSON(400, gin.H{
                "status": false,
                "err_message": "當提供satisfyQuality時，至少必須提供一個satisfyQuality.slaProfile項目",
            })
            return
        }
    }

    // 使用互斥鎖保護共享資源profileStore以防止同時讀寫
    profileMutex.Lock()
    defer profileMutex.Unlock()

    // 檢查存儲中是否存在指定的profile
    profile, exists := profileStore[request.ProfileID]
    if !exists {
        c.JSON(404, gin.H{"status": false, "err_message": "找不到Profile"})
        return
    }

    // 使用請求中的已驗證數據更新profile
    if request.Name != "" {
        profile.Name = request.Name
    }
    if request.BestQuality != nil {
        profile.BestQuality = request.BestQuality
    }
    if request.SatisfyQuality != nil {
        profile.SatisfyQuality = request.SatisfyQuality
    }

    // 回傳成功的響應
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
            indicator = getIndicatorFromSLAProfile(profile.SatisfyQuality.SLAProfile)
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

func getIndicatorFromSLAProfile(sla *SLAProfile) string {
    if sla.BandwidthOptimized != nil {
        return "bandwidthOptimized"
    }
    if sla.LatencySensitive != nil {
        return "latencySensitive"
    }
    if sla.QualityStability != nil {
        return "qualityStability"
    }
    if sla.ReliabilityMonitor != nil {
        return "reliabilityMonitor"
    }
    return ""
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
