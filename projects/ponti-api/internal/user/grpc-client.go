package user

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	sdkgrpc "github.com/devpablocristo/sdk/pkg/grpc/client"

// 	pb "github.com/devpablocristo/sdk/pb"
// 	ports "github.com/devpablocristo/sdk/sg/users/internal/core/ports"
// )

// // GrpcClient estructura que representa el cliente gRPC para AuthService
// type GrpcClient struct {
// 	authClient pb.AuthServiceClient
// }

// // NewGrpcClient crea una nueva instancia de GrpcClient
// func NewGrpcClient() (ports.GrpcClient, error) {

// 	// Inicializa el cliente gRPC usando el SDK
// 	client, err := sdkgrpc.Bootstrap("AUTH_GRPC_SERVER_HOST", "AUTH_GRPC_SERVER_PORT")
// 	if err != nil {
// 		return nil, fmt.Errorf("bootstrap error: %w", err)
// 	}

// 	// Obtiene la conexión gRPC desde el cliente inicializado
// 	conn, err := client.GetConnection()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get gRPC connection: %w", err)
// 	}

// 	// Crea un cliente para el servicio AuthService
// 	authClient := pb.NewAuthServiceClient(conn)

// 	// Retorna una instancia de GrpcClient que contiene el cliente AuthService
// 	return &GrpcClient{
// 		authClient: authClient,
// 	}, nil
// }

// // RequestVerificationToken solicita a auth un token de verificación
// func (g *GrpcClient) RequestVerificationToken(ctx context.Context, userID string) (string, string, error) {
// 	// Configura un contexto con tiempo de espera
// 	c, cancel := context.WithTimeout(ctx, 1*time.Minute)
// 	defer cancel()

// 	// Crea la solicitud con el UUID del usuario
// 	req := &pb.VerificationTokenRequest{
// 		UserUuid: userID,
// 	}

// 	// Llama al método GenerateVerificationToken en auth y maneja la respuesta
// 	res, err := g.authClient.GenerateTokens(c, req)
// 	if err != nil {
// 		return "", "", fmt.Errorf("failed to request verification token: %w", err)
// 	}

// 	// Devuelve el token de verificación recibido
// 	return res.AccessToken, res.RefreshToken, nil
// }

// // Close cierra la conexión gRPC (si es necesario)
// func (g *GrpcClient) Close() error {
// 	// Si necesitas cerrar la conexión, llama a conn.Close() aquí, si tienes acceso a ella.
// 	return nil
// }
