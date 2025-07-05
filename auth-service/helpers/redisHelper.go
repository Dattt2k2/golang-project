package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"auth-service/database"
)


type DeviceInfo struct{
    DeviceID string `json:"device_id"`
    Platfrom string `json:"platfrom"`
    LastLoginAt time.Time `json:"last_login_at"`
    IPAddress string `json:"ip_address"`
    UserAgent string `json:"user_agent"`
}

func StoreRefreshToken(userId, refreshToken, deviceId, platfrom, ipAddress, userAgent string) error {
	ctx := context.Background()

	key := fmt.Sprintf("refresh_token:%s", userId)
	if deviceId != ""{
		key = fmt.Sprint("refresh_token:", userId, deviceId)
	}

	deviceInfo := DeviceInfo{
		DeviceID: deviceId,
		Platfrom: platfrom,
		LastLoginAt: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	deviceJSON, err := json.Marshal(deviceInfo)
	if err != nil{
		return err
	}

	pipe := database.RedisClient.Pipeline()

	pipe.Set(ctx, key, refreshToken, 72*time.Hour)

	deviceInfoKey := fmt.Sprintf("device_info:%s:%s", userId, deviceId)
	pipe.Set(ctx, deviceInfoKey, string(deviceJSON), 7*24*time.Hour)

	userDevicesKey := fmt.Sprintf("user_devices:%s", userId)
	pipe.SAdd(ctx, userDevicesKey, deviceId)
	pipe.Expire(ctx, userDevicesKey, 30*24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}



func GetRefreshToken(userId, deviceId string) (string, error ){
	ctx := context.Background()

	key := fmt.Sprintf("refresh_token:%s", userId)
	if deviceId != ""{
		key = fmt.Sprintf("refresh_token:%s:%s", userId, deviceId)
	}

	return database.RedisClient.Get(ctx, key).Result()
}


func InvalidateRefreshToken(userId, deviceId string) error {
	ctx := context.Background()

	if deviceId == ""{
		return fmt.Errorf("device ID is required")
	}

	key := fmt.Sprintf("refresh_token:%s:%s", userId, deviceId)

	deviceInfoKey := fmt.Sprintf("device_info:%s:%s", userId, deviceId)

	userDevicesKey := fmt.Sprintf("user_devices:%s", userId)

	pipe := database.RedisClient.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, deviceInfoKey)
	pipe.SRem(ctx, userDevicesKey, deviceId)
	_, err := pipe.Exec(ctx)

	return err

}

func InvalidateAllUserRefreshToken(userId string) error {
	ctx := context.Background()

	userDevicesKey := fmt.Sprintf("user_devices:%s", userId)
	deviceIds, err := database.RedisClient.SMembers(ctx, userDevicesKey).Result()
	if err != nil{
		return err
	}

	pipe := database.RedisClient.Pipeline()

	for _, deviceId := range deviceIds{
		tokenKey := fmt.Sprintf("refresh_token:%s:%s", userId, deviceId)
		deviceInfoKey := fmt.Sprintf("device_info:%s:%s", userId, deviceId)

		pipe.Del(ctx, tokenKey)
		pipe.Del(ctx, deviceInfoKey)
	}

	pipe.Del(ctx, userDevicesKey)

	_, err = pipe.Exec(ctx)

	return err
}


func GetUserDevices(userId string) ([]DeviceInfo, error){
	ctx := context.Background()

	userDeviceKey := fmt.Sprintf("user_devices:%s", userId)
	deviceIds, err := database.RedisClient.SMembers(ctx, userDeviceKey).Result()

	if err != nil{
		return nil, err
	}

	devices := make([]DeviceInfo, 0, len(deviceIds))

	for _, deviceId := range deviceIds {
		deviceInfoKey := fmt.Sprintf("device_info:%s:%s", userId, deviceId)
		deviceInfo, err := database.RedisClient.Get(ctx, deviceInfoKey).Result()
		if err != nil{
			return nil, err
		}

		var device DeviceInfo
		err = json.Unmarshal([]byte(deviceInfo), &device)
		if err != nil{
			return nil, err
		}

		devices = append(devices, device)
	}

	return devices, nil
}


func isDeviceTrusted(userId, deviceId string) (bool, error){
	if deviceId == ""{
		return false, nil
	}

	ctx := context.Background()
	userDevicesKey := fmt.Sprintf("user_devices:%s", userId)

	return database.RedisClient.SIsMember(ctx, userDevicesKey, deviceId).Result()
}


func VerifyRefreshToken(userId, deviceId, providedToken string) (bool, error) {
	if providedToken == "" || userId == ""{
		return false, fmt.Errorf("invalid token or user ID")
	}

	storedToken, err := GetRefreshToken(userId, deviceId)
	if err != nil{
		return false, err
	}

	return storedToken == providedToken, nil
}


func UpdateDeviceLastLogin(userId, deviceId, ipAddress string) error {
	ctx := context.Background()

	deviceInfoKey := fmt.Sprintf("device_info:%s:%s", userId, deviceId)

	deviceJSON, err := database.RedisClient.Get(ctx, deviceInfoKey).Result()

	if err != nil{
		return err
	}

	var deviceInfo DeviceInfo
	if err := json.Unmarshal([]byte(deviceJSON), &deviceInfo); err != nil{
		return err
	}

	deviceInfo.LastLoginAt = time.Now()
	if ipAddress != ""{
		deviceInfo.IPAddress = ipAddress
	}

	updateJSON, err := json.Marshal(deviceInfo)
	if err != nil{
		return err
	}

	return database.RedisClient.Set(ctx, deviceInfoKey, string(updateJSON), 7*24*time.Hour).Err()
}