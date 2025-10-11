package config

import (
    "os"
    "strconv"
    "strings"

    "github.com/joho/godotenv"
)


func Load(paths ...string) error {
    if len(paths) == 0 {
        _ = godotenv.Load()
        return nil
    }
    for _, p := range paths {
        _ = godotenv.Overload(p)
    }
    return nil
}

func Get(key, defaultValue string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultValue
}

func MustGet(key string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    panic("missing required env: " + key)
}

func GetInt(key string, defaultValue int) int {
    if v := os.Getenv(key); v != "" {
        if i, err := strconv.Atoi(v); err == nil {
            return i
        }
    }
    return defaultValue
}

func GetBool(key string, defaultValue bool) bool {
    if v := os.Getenv(key); v != "" {
        l := strings.ToLower(v)
        if l == "1" || l == "true" || l == "yes" {
            return true
        }
        return false
    }
    return defaultValue
}

func SplitAndTrim(value string, sep string) []string {
    if value == "" {
        return nil
    }
    parts := strings.Split(value, sep)
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        if s := strings.TrimSpace(p); s != "" {
            out = append(out, s)
        }
    }
    return out
}