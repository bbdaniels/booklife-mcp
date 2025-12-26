package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/user/booklife-mcp/internal/models"
)

// registerPrompts registers all BookLife prompts with the MCP server
func (s *Server) registerPrompts() {
	// What should I read next?
	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "what_should_i_read",
		Description: "Get a personalized book recommendation based on your mood and preferences",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "mood",
				Description: "Your current reading mood (e.g., 'something light', 'deep and thoughtful', 'fast-paced')",
				Required:    false,
			},
			{
				Name:        "time",
				Description: "How much time you have (e.g., 'quick read', 'weekend', 'vacation')",
				Required:    false,
			},
			{
				Name:        "format",
				Description: "Preferred format: ebook, audiobook, physical, or any",
				Required:    false,
			},
		},
	}, s.handleWhatShouldIRead)

	// Book summary
	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "book_summary",
		Description: "Get a comprehensive summary of a book without spoilers",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "title",
				Description: "Book title",
				Required:    true,
			},
			{
				Name:        "spoiler_level",
				Description: "How much to reveal: none, light, or full",
				Required:    false,
			},
		},
	}, s.handleBookSummary)

	// Reading wrap-up
	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "reading_wrap_up",
		Description: "Generate a reading wrap-up for a time period",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "period",
				Description: "Time period: week, month, or year",
				Required:    true,
			},
		},
	}, s.handleReadingWrapUp)

	// Compare books
	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "compare_books",
		Description: "Compare two or more books side by side",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "books",
				Description: "Comma-separated book titles or ISBNs to compare",
				Required:    true,
			},
			{
				Name:        "aspects",
				Description: "What to compare: themes, writing style, pacing, characters, or all",
				Required:    false,
			},
		},
	}, s.handleCompareBooks)

	// Decide what to read from TBR
	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "pick_from_tbr",
		Description: "Help decide what to read next from your TBR pile",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "constraints",
				Description: "Any constraints like 'available at library' or 'under 300 pages'",
				Required:    false,
			},
		},
	}, s.handlePickFromTBR)
}

// === Prompt Handlers ===

// formatBookLine formats a single book for display in prompts.
func formatBookLine(book models.Book, includeRating bool) string {
	var b strings.Builder
	b.WriteString("- ")
	b.WriteString(book.Title)
	if len(book.Authors) > 0 {
		b.WriteString(" by ")
		b.WriteString(book.Authors[0].Name)
	}
	if includeRating && book.UserStatus != nil && book.UserStatus.Rating > 0 {
		b.WriteString(fmt.Sprintf(" (rated %.1f/5)", book.UserStatus.Rating))
	}
	if book.PageCount > 0 {
		b.WriteString(fmt.Sprintf(" (%d pages)", book.PageCount))
	}
	b.WriteString("\n")
	return b.String()
}

// buildHardcoverContext builds context string from Hardcover reading data.
func (s *Server) buildHardcoverContext(ctx context.Context) string {
	if s.hardcover == nil {
		return ""
	}

	var b strings.Builder

	// Get recently finished books
	recent, err := s.hardcover.GetUserBooks(ctx, "read", 10)
	if err == nil && len(recent) > 0 {
		b.WriteString("Recently finished books:\n")
		for _, book := range recent {
			b.WriteString(formatBookLine(book, true))
		}
	}

	// Get TBR
	tbr, err := s.hardcover.GetUserBooks(ctx, "want-to-read", 20)
	if err == nil && len(tbr) > 0 {
		b.WriteString("\nBooks on TBR:\n")
		for _, book := range tbr {
			b.WriteString(formatBookLine(book, false))
		}
	}

	return b.String()
}

// buildLibbyContext builds context string from Libby library data.
func (s *Server) buildLibbyContext(ctx context.Context) string {
	if s.libby == nil {
		return ""
	}

	var b strings.Builder

	loans, err := s.libby.GetLoans(ctx)
	if err == nil && len(loans) > 0 {
		b.WriteString("\nCurrently borrowed from library:\n")
		for _, l := range loans {
			b.WriteString(fmt.Sprintf("- %s (due: %s)\n", l.Title, l.DueDate.Format("Jan 2")))
		}
	}

	holds, err := s.libby.GetHolds(ctx)
	if err == nil && len(holds) > 0 {
		b.WriteString("\nHolds ready or coming soon:\n")
		for _, h := range holds {
			if h.IsReady {
				b.WriteString(fmt.Sprintf("- %s (READY NOW!)\n", h.Title))
			} else {
				b.WriteString(fmt.Sprintf("- %s (~%d day wait)\n", h.Title, h.EstimatedWaitDays))
			}
		}
	}

	return b.String()
}

// buildRecommendationPrompt builds the system and user prompts for recommendations.
func buildRecommendationPrompt(mood, timeAvailable, format, libraryContext, availabilityContext string) string {
	systemPrompt := `You are a knowledgeable and enthusiastic book recommendation assistant.
Your goal is to help the reader find their perfect next read based on their mood, preferences, and what's available to them.

Consider:
1. Their reading history and what they've enjoyed
2. Their current mood and time constraints
3. What's available at their library (free is always best!)
4. Books on their TBR that match their current mood

Be specific with recommendations and explain WHY each book would be a good fit.
Prioritize books that are available for free at the library when possible.`

	var userPrompt strings.Builder
	userPrompt.WriteString("I'm looking for my next read.\n")
	if mood != "" {
		userPrompt.WriteString(fmt.Sprintf("Mood: %s\n", mood))
	}
	if timeAvailable != "" {
		userPrompt.WriteString(fmt.Sprintf("Time available: %s\n", timeAvailable))
	}
	if format != "" {
		userPrompt.WriteString(fmt.Sprintf("Preferred format: %s\n", format))
	}
	if libraryContext != "" {
		userPrompt.WriteString("\n--- My Reading History ---\n")
		userPrompt.WriteString(libraryContext)
	}
	if availabilityContext != "" {
		userPrompt.WriteString("\n--- Library Availability ---\n")
		userPrompt.WriteString(availabilityContext)
	}
	userPrompt.WriteString("\nWhat should I read next?")

	return systemPrompt + "\n\n" + userPrompt.String()
}

func (s *Server) handleWhatShouldIRead(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	mood := req.Params.Arguments["mood"]
	timeAvailable := req.Params.Arguments["time"]
	format := req.Params.Arguments["format"]

	libraryContext := s.buildHardcoverContext(ctx)
	availabilityContext := s.buildLibbyContext(ctx)

	prompt := buildRecommendationPrompt(mood, timeAvailable, format, libraryContext, availabilityContext)

	return &mcp.GetPromptResult{
		Description: "Personalized book recommendation",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: prompt,
				},
			},
		},
	}, nil
}

func (s *Server) handleBookSummary(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	title := req.Params.Arguments["title"]
	spoilerLevel := req.Params.Arguments["spoiler_level"]
	if spoilerLevel == "" {
		spoilerLevel = "none"
	}

	var bookContext string
	// Try to get book metadata
	if s.hardcover != nil {
		books, err := s.hardcover.SearchBooks(ctx, title, 1)
		if err == nil && len(books) > 0 {
			book := books[0]
			bookContext = fmt.Sprintf(`
Book: %s
Author: %s
Genres: %v
Description: %s
`, book.Title, book.Authors[0].Name, book.Genres, book.Description)
		}
	}

	systemPrompt := fmt.Sprintf(`You are a book summarizer. Provide a summary of the requested book.
Spoiler level: %s
- none: Only premise and themes, no plot spoilers
- light: General plot arc without major twists
- full: Complete summary including ending`, spoilerLevel)

	userPrompt := fmt.Sprintf("Please summarize: %s\n%s", title, bookContext)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Summary of %s", title),
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: systemPrompt + "\n\n" + userPrompt,
				},
			},
		},
	}, nil
}

func (s *Server) handleReadingWrapUp(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	period := req.Params.Arguments["period"]

	var readingData string
	if s.hardcover != nil {
		// Get books read in period
		books, _ := s.hardcover.GetUserBooks(ctx, "read", 100)
		if len(books) > 0 {
			readingData = "Books read:\n"
			for _, b := range books {
				readingData += fmt.Sprintf("- %s by %s", b.Title, b.Authors[0].Name)
				if b.UserStatus != nil && b.UserStatus.Rating > 0 {
					readingData += fmt.Sprintf(" (%.1f/5)", b.UserStatus.Rating)
				}
				readingData += "\n"
			}
		}
	}

	systemPrompt := `You are creating a reading wrap-up. Analyze the books read and create an engaging summary that includes:
1. Total books read
2. Favorite reads (highest rated)
3. Genre breakdown
4. Themes across the reading
5. A memorable quote or moment
6. Looking ahead - what patterns suggest for future reading`

	userPrompt := fmt.Sprintf("Create a %s reading wrap-up.\n\n%s", period, readingData)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("%s reading wrap-up", period),
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: systemPrompt + "\n\n" + userPrompt,
				},
			},
		},
	}, nil
}

func (s *Server) handleCompareBooks(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	books := req.Params.Arguments["books"]
	aspects := req.Params.Arguments["aspects"]
	if aspects == "" {
		aspects = "all"
	}

	systemPrompt := fmt.Sprintf(`You are comparing books. Compare the requested books on these aspects: %s

For each aspect, provide:
- How each book approaches it
- Which book excels
- Who would prefer which

Be balanced and help the reader understand the differences to make an informed choice.`, aspects)

	userPrompt := fmt.Sprintf("Compare these books: %s", books)

	return &mcp.GetPromptResult{
		Description: "Book comparison",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: systemPrompt + "\n\n" + userPrompt,
				},
			},
		},
	}, nil
}

func (s *Server) handlePickFromTBR(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	constraints := req.Params.Arguments["constraints"]

	var tbrContext string
	if s.hardcover != nil {
		tbr, _ := s.hardcover.GetUserBooks(ctx, "want-to-read", 50)
		if len(tbr) > 0 {
			tbrContext = "TBR pile:\n"
			for _, b := range tbr {
				tbrContext += fmt.Sprintf("- %s by %s", b.Title, b.Authors[0].Name)
				if b.PageCount > 0 {
					tbrContext += fmt.Sprintf(" (%d pages)", b.PageCount)
				}
				tbrContext += "\n"
			}
		}
	}

	// Check which TBR books are available at library
	var availabilityContext string
	if s.libby != nil {
		availabilityContext = "\nLibrary availability for TBR books:\n"
		// TODO: Check each TBR book against library
	}

	systemPrompt := `You are helping someone choose what to read from their TBR pile.

Consider:
1. Any constraints they mentioned
2. What's available at the library (free!)
3. Variety - if they've read a lot of one genre recently, suggest something different
4. Book length vs available time

Narrow it down to 2-3 strong recommendations with clear reasoning.`

	userPrompt := fmt.Sprintf("Help me pick from my TBR.\n%s\n%s", tbrContext, availabilityContext)
	if constraints != "" {
		userPrompt += fmt.Sprintf("\nConstraints: %s", constraints)
	}

	return &mcp.GetPromptResult{
		Description: "TBR picker",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: systemPrompt + "\n\n" + userPrompt,
				},
			},
		},
	}, nil
}
