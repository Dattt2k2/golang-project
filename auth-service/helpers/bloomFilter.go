package helpers

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"time"

	"golang.org/x/net/context"
	"gorm.io/gorm"

	database "auth-service/database"
	"auth-service/models"

	"github.com/redis/go-redis/v9"
)


type BloomFilter struct{
	Name string
	M uint
	K uint
}

const (
	DefaultM = 100000

	DefaultK = 5
)

func NewBloomFilter(name string, m, k uint) *BloomFilter{
	if m == 0{
		m = DefaultM
	}
	if k == 0{
		k = DefaultK
	}
	return &BloomFilter{
		Name: name,
		M: m,
		K: k,
	}
}

func NewUserBloomFilter() *BloomFilter{
	return NewBloomFilter("user_bloom", DefaultM, DefaultK)
}

func (bf *BloomFilter) Init(db *gorm.DB) error{
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database.RedisClient.Del(ctx, bf.Name)

	var users []struct {
        Email   string
        Username string
        Phone   string
    }
    // Nếu bạn không có trường Username thì bỏ Username đi
    if err := db.Model(&models.User{}).Select("email, phone").Find(&users).Error; err != nil {
        return err
    }

	pipe := database.RedisClient.Pipeline()
	count := 0

	for _, user := range users {
		if user.Email != "" {
			bf.addToPipeline(pipe, user.Email)
			count++
		}
		if user.Phone != "" {
			bf.addToPipeline(pipe, user.Phone)
			count++
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (bf *BloomFilter) addToPipeline(pipe redis.Pipeliner, value string){
	positions := bf.positions(value)

	for _, pos := range positions{
		pipe.SetBit(context.Background(), bf.Name, int64(pos), 1)
	}

}

func (bf *BloomFilter) Add(value string) error{
	ctx := context.Background()

	positions := bf.positions(value)

	pipe := database.RedisClient.Pipeline()

	for _, pos := range positions{
		pipe.SetBit(ctx, bf.Name, int64(pos), 1)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (bf *BloomFilter) Contains(value string) (bool, error){


	ctx := context.Background()
	posistions := bf.positions(value)

	pipe := database.RedisClient.Pipeline()

	results := make([]*redis.IntCmd, len(posistions))
	for i, pos := range posistions{
		results[i] = pipe.GetBit(ctx, bf.Name, int64(pos))
	}

	_, err := pipe.Exec(ctx)
	if err != nil{
		return false, err
	}

	for _, result := range results{
		bit, err := result.Result()
		if err != nil{
			return false, err
		}
		if bit == 0{
			return false, nil
		}
	}

	return true, nil
}

func (bf *BloomFilter) positions(value string) []uint{
	positions := make([]uint, bf.K)

	h1 := sha1.Sum([]byte(value))
	h2 := sha256.Sum256([]byte(value))

	a := binary.BigEndian.Uint32(h1[0:4])
	b := binary.BigEndian.Uint32(h2[0:4])

	for i := uint(0); i < bf.K; i++{
		positions[i] = uint((uint64(a) + uint64(i) * uint64(b)) % uint64(bf.M))
	}

	return positions
}

func CalculateOptimalSize(n int, p float64)  (m uint, k uint){

    m = uint(math.Ceil(-float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))

	k = uint(math.Ceil(float64(m) / float64(n) * math.Log(2)))
    
    return m, k
}


func (bf *BloomFilter) EstimateFalsePositiveRate(n int) float64{
	
	k := float64(bf.K)
	m := float64(bf.M)

	return math.Pow(1 - math.Exp(-k * float64(n) / m), k)
}

func CreateOptimalUserBloomFilter(expectedUsers int) *BloomFilter{
	m, k := CalculateOptimalSize(expectedUsers, 0.001)
	return NewBloomFilter("user_bloom", m, k)
}




