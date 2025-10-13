package auth

import (
    "errors"
    "time"
    "os"

    "github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte(getJwtKey())

func getJwtKey() string {
    k := os.Getenv("JWT_KEY")
    if k == "" {
        return "verysecret"
    }
    return k
}

type Claims struct {
    UserID int64  `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID int64, role string) (string, error) {
    claims := &Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}

func ParseToken(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, errors.New("invalid token")
}
