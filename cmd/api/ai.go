package main

import (
	"context"
	"google.golang.org/genai"
	"net/http"
	"strings"
)

func (app *application) classifyReceiptHandler(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := app.readImage(r, 2<<20, "receipt") // Image can have a max size of 2 MB and should be sent with the "receipt" key
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: app.config.gemini.apiKey,
	})

	// The first text is the prompt for Gemini which is then executed with the image to classify the expenses
	parts := []*genai.Part{
		{Text: "You are an expense classification expert. You have to look at the receipt from a grocery store and calculate the total amount of money spent on:\nGroceries: milk, fruits, vegetables, or anything else that can be used to make food at home.\nSnacks:  junk food such as chips, popcorn, crackers, frozen pizza, or anything unhealthy.\nDessert: anything with added sugar such as chocolates, ice cream, cheesecakes, brownies, etc.\nCoffee: coffee beans, instant coffee, etc.\n\nDo not make any mistake and make sure all amounts add up to the final amount on the receipt.\n\nAfter calculating the total money spent on the 4 categories, you have to give your response in the following JSON schema:\nresult: Array of JSON object:\n{\namount\nspent_at (name of the grocery store)\ncategory\npayment_method (should be on the receipt, else leave it blank)\n}\nLeave out any categories where items were not purchases. Do not repeat any categories. Each category should have the total of all the items purchases in that category."},
		{InlineData: &genai.Blob{Data: fileBytes, MIMEType: "image/jpeg"}},
	}
	result, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash", []*genai.Content{{Parts: parts}}, nil)

	text, err := result.Text()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// The response from Gemini starts with ```json and ends with ```, so we trim that from the response and returning it without those markers
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimSuffix(text, "```")

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(text))
}
