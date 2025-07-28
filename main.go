package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	ctx := context.TODO()

	// text := `
	// You are my sunshine, my only sunshine.
	// You make me happy when skies are gray.
	// You'll never know, dear, how much I love you.
	// Please don't take my sunshine away.`
	textEsp := `Eres mi sol, mi único sol.
	Me haces feliz cuando el cielo está gris.
	Nunca sabrás, querida, cuánto te amo.
	Por favor, no me quites mi sol.`
	bucketName := "tin-buckets-and-plates"
	objectKey := "polly/streamed-output-neural-es.mp3"

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	// Polly client
	pollyClient := polly.NewFromConfig(cfg)

	// Synthesize speech
	speechResp, err := pollyClient.SynthesizeSpeech(ctx, &polly.SynthesizeSpeechInput{
		Text:         aws.String(textEsp),
		OutputFormat: "mp3",
		VoiceId:      "Lupe",
		Engine:       "neural",
		LanguageCode: "es-US", // Spanish (United States)
	})
	if err != nil {
		log.Fatalf("failed to synthesize speech: %v", err)
	}
	defer speechResp.AudioStream.Close()

	// Read stream into memory
	var buf bytes.Buffer
	size, err := io.Copy(&buf, speechResp.AudioStream)
	if err != nil {
		log.Fatalf("failed to buffer audio stream: %v", err)
	}

	// Upload to S3
	s3Client := s3.NewFromConfig(cfg)

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(objectKey),
		Body:          bytes.NewReader(buf.Bytes()),
		ContentLength: &size,
		ContentType:   aws.String("audio/mpeg"),
	})
	if err != nil {
		log.Fatalf("failed to upload to S3: %v", err)
	}

	fmt.Printf("Audio uploaded to s3://%s/%s (%d bytes)\n", bucketName, objectKey, size)
}
