package service

import (
	"fmt"
	"planA/planB/initialization/golabl"
	"strconv"
)

// GetRegionId 根据地区Id获取获取省市信息
// @param districtID 地区ID
// @return int 省份ID
// @return int 城市ID
// @return int 区县ID
// @return error 错误信息
func GetRegionId(districtID string) (int, int, int, error) {
	//获取区县 code
	district, err := golabl.Redis.RedisDbC.HGetAll(golabl.Ctx, fmt.Sprintf("region:%s", districtID)).Result()
	if err != nil {
		return 0, 0, 0, err
	}
	// 将 district["code"] 转为 int
	districtCode, districtErr := strconv.Atoi(district["code"])
	if districtErr != nil {
		return 0, 0, 0, fmt.Errorf("区县code转换失败 id %v %v", districtID, districtErr)
	}
	//获取市 code
	city, err := golabl.Redis.RedisDbC.HGetAll(golabl.Ctx, fmt.Sprintf("region:%s", district["pid"])).Result()
	if err != nil {
		return 0, 0, 0, err
	}
	// 将 city["code"] 转为 int
	cityCode, cityErr := strconv.Atoi(city["code"])
	if cityErr != nil {
		return 0, 0, 0, fmt.Errorf("市code转换失败 id %v %v", district["pid"], err)
	}
	//获取市 province
	province, err := golabl.Redis.RedisDbC.HGetAll(golabl.Ctx, fmt.Sprintf("region:%s", city["pid"])).Result()
	if err != nil {
		return 0, 0, 0, err
	}
	// 将 province["code"] 转为 int
	provinceCode, provinceErr := strconv.Atoi(province["code"])
	if provinceErr != nil {
		return 0, 0, 0, fmt.Errorf("省code转换失败 id %v %v", city["pid"], err)
	}
	return provinceCode, cityCode, districtCode, nil
}

// GetRandomDistrictInProvince 在指定省内随机获取一个区级地区
func GetRandomDistrictInProvince(provinceID int) (map[string]string, error) {
	// 从该省份的区级地区集合中随机获取一个 ID
	provinceKey := fmt.Sprintf("province:%d:districts", provinceID)
	districtID, err := golabl.Redis.RedisDbC.SRandMember(golabl.Ctx, provinceKey).Result()
	if err != nil {
		return nil, err
	}

	// 获取该地区的详细信息
	return golabl.Redis.RedisDbC.HGetAll(golabl.Ctx, fmt.Sprintf("region:%s", districtID)).Result()
}

// GetRandomDistrict 随机获取一个区级地区
func GetRandomDistrict() (map[string]string, error) {
	// 从所有区级地区集合中随机获取一个 ID
	districtID, err := golabl.Redis.RedisDbC.SRandMember(golabl.Ctx, "all:districts").Result()
	if err != nil {
		return nil, err
	}

	// 获取该地区的详细信息
	return golabl.Redis.RedisDbC.HGetAll(golabl.Ctx, fmt.Sprintf("region:%s", districtID)).Result()
}
