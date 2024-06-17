// package utils

// import (
// 	"challenge/database"
// 	"challenge/entity"
// 	"context"

// 	"github.com/gofiber/fiber/v2"
// 	jtoken "github.com/golang-jwt/jwt/v4"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// func GetUserByToken(ctx context.Context, db *database.DB) (*entity.CustomerEntity, error) {
// 	token := ctx.Value("user").(*jtoken.Token)
// 	claims, ok := token.Claims.(jtoken.MapClaims)
// 	if !ok {
// 		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to parse token claims")
// 	}
// 	userIDHex, ok := claims["Id"].(string)
// 	if !ok {
// 		return nil, fiber.NewError(fiber.StatusInternalServerError, "UserId not found in token claims")
// 	}

// 	userID, err := primitive.ObjectIDFromHex(userIDHex)
// 	if err != nil {
// 		return nil, fiber.NewError(fiber.StatusInternalServerError, "Invalid user ID in token claims")
// 	}

//		customerColl := db.GetCollection("user")
//		var user entity.CustomerEntity
//		err = customerColl.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
//		if err != nil {
//			return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch user details: "+err.Error())
//		}
//		return &user, nil
//	}
package utils

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type ContextKey string

const UserClaimsKey ContextKey = "userClaims"

// ExtractClaimsFromContext retrieves JWT claims from the context
func ExtractClaimsFromContext(ctx context.Context) (jwt.MapClaims, error) {
	claims, ok := ctx.Value(UserClaimsKey).(jwt.MapClaims)
	if !ok {
		fmt.Println("No claims found in context") // Debugging log for missing claims
		return nil, fmt.Errorf("no claims found in context")
	}
	fmt.Println("Claims extracted from context: %v", claims) // Debugging log for extracted claims
	return claims, nil
}

// ExtractUserIDFromContext retrieves the user ID from the JWT claims in the context
func ExtractUserIDFromContext(ctx context.Context) (string, error) {
	claims, err := ExtractClaimsFromContext(ctx)
	if err != nil {
		return "", err
	}
	userID, ok := claims["Id"].(string)
	if !ok {
		fmt.Println("User ID not found in token claims") // Debugging log for missing user ID
		return "", fmt.Errorf("user ID not found in token claims")
	}
	fmt.Println("User ID extracted: %s", userID) // Debugging log for extracted user ID
	return userID, nil
}

// ExtractDeviceIDFromContext retrieves the device ID from the JWT claims in the context
func ExtractDeviceIDFromContext(ctx context.Context) (string, error) {
	claims, err := ExtractClaimsFromContext(ctx)
	if err != nil {
		return "", err
	}
	deviceID, ok := claims["deviceId"].(string)
	if !ok {
		fmt.Println("Device ID not found in token claims") // Debugging log for missing device ID
		return "", fiber.NewError(fiber.StatusInternalServerError, "DeviceId not found in token claims")
	}
	fmt.Println("Device ID extracted: %s", deviceID) // Debugging log for extracted device ID
	return deviceID, nil
}
