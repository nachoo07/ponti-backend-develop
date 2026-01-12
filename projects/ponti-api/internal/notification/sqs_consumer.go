package notification

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"

// 	"github.com/devpablocristo/customer-manager/pkg/aws/localstack/defs"
// 	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/notification/handler/dto"
// )

// const queueName = "users_verification_email"

// // SQSConsumer representa el consumidor de SQS para notification.
// type SQSConsumer struct {
// 	sqsClient defs.SqsClient
// 	queueURL  string
// 	ucs       ports.UseCases
// }

// // NewSQSConsumer crea una nueva instancia de SQSConsumer.
// // Se inyecta la dependencia de UseCases para que el consumidor pueda invocar la lógica de negocio.
// func NewSQSConsumer(u ports.UseCases) (ports.SqsConsumer, error) {
// 	// Inicializa el stack de AWS.
// 	stack, err := sdkaws.Bootstrap()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to initialize AWS stack: %w", err)
// 	}

// 	// Obtén el cliente SQS.
// 	sqsClient := stack.NewSQSClient()

// 	// Obtén o crea la cola usando el SDK.
// 	queueURL, err := sqsClient.GetOrCreateQueueURL(context.TODO(), queueName)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get or create queue '%s': %w", queueName, err)
// 	}

// 	return &SQSConsumer{
// 		sqsClient: sqsClient,
// 		queueURL:  queueURL,
// 		ucs:       u,
// 	}, nil
// }

// // StartConsuming inicia el proceso de consumo de mensajes.
// func (q *SQSConsumer) StartConsuming(ctx context.Context) error {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return nil
// 		default:
// 			messages, err := q.sqsClient.ReceiveMessages(ctx, q.queueURL, 10)
// 			if err != nil {
// 				fmt.Printf("Error receiving messages: %v\n", err)
// 				continue
// 			}

// 			for _, message := range messages {
// 				if err := q.processMessage(ctx, message); err != nil {
// 					fmt.Printf("Error processing message: %v\n", err)
// 					// Aquí podrías implementar lógica para manejar mensajes fallidos, como enviarlos a una DLQ.
// 					continue
// 				}

// 				// Eliminar el mensaje después de procesarlo correctamente.
// 				if err := q.sqsClient.DeleteMessage(ctx, q.queueURL, message.ReceiptHandle); err != nil {
// 					fmt.Printf("Error deleting message: %v\n", err)
// 				}
// 			}
// 		}
// 	}
// }

// // processMessage procesa un mensaje recibido.
// func (q *SQSConsumer) processMessage(ctx context.Context, message defs.SQSMessage) error {
// 	var verificationMsg dto.VerificationMessage
// 	if err := json.Unmarshal([]byte(message.Body), &verificationMsg); err != nil {
// 		return fmt.Errorf("error unmarshalling message body: %w", err)
// 	}

// 	// Enviar el correo electrónico de verificación utilizando la lógica de negocio.
// 	if err := q.ucs.SendVerificationEmail(ctx, verificationMsg.Email, verificationMsg.Token); err != nil {
// 		return fmt.Errorf("error sending verification email: %w", err)
// 	}

// 	return nil
// }
