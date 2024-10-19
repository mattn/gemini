package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type payload struct {
	Text string `json:"text"`
}

type response struct {
	EOF   bool   `json:"eof"`
	Error string `json:"error"`
	Text  string `json:"text"`
}

func doJson(model *genai.GenerativeModel, r io.Reader, w io.Writer) error {
	enc := json.NewEncoder(w)
	dec := json.NewDecoder(r)
	ctx := context.Background()
	for {
		var p payload
		err := dec.Decode(&p)
		if err != nil {
			return err
		}

		iter := model.GenerateContentStream(ctx, genai.Text(p.Text))
		for {
			resp, err := iter.Next()
			if err != nil {
				break
			}
			for _, c := range resp.Candidates {
				for _, p := range c.Content.Parts {
					err = enc.Encode(response{Text: fmt.Sprint(p)})
					if err != nil {
						return err
					}
				}
			}
		}

		err = enc.Encode(response{EOF: true})
		if err != nil {
			return err
		}
	}
}
func main() {
	var j bool
	var m string
	flag.StringVar(&m, "model", "gemini-1.5-flash", "model name")
	flag.BoolVar(&j, "json", false, "json input/output")
	flag.Parse()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing GEMINI_API_KEY")
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(m)

	if j {
		log.Fatal(doJson(model, os.Stdin, os.Stdout))
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		iter := model.GenerateContentStream(ctx, genai.Text(text))
		for {
			resp, err := iter.Next()
			if err != nil {
				break
			}
			for _, c := range resp.Candidates {
				for _, p := range c.Content.Parts {
					fmt.Print(p)
				}
			}
		}
		fmt.Println()
	}
}
