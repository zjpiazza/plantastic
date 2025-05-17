package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/zjpiazza/plantastic/internal/models"
)

func gardensCmd(apiUrl string) *cobra.Command {
	gardensCmd := &cobra.Command{
		Use:   "gardens",
		Short: "Manage your gardens",
		Long:  `Create, list, update, and delete your gardens.`,
	}

	gardensCmd.AddCommand(listGardensCmd(apiUrl))
	gardensCmd.AddCommand(createGardenCmd(apiUrl))
	gardensCmd.AddCommand(updateGardenCmd(apiUrl))
	gardensCmd.AddCommand(deleteGardenCmd(apiUrl))

	return gardensCmd
}

func listGardensCmd(apiUrl string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all gardens",
		Run: func(cmd *cobra.Command, args []string) {
			response, err := http.Get(apiUrl + "/gardens")
			if err != nil {
				fmt.Println("Error getting gardens:", err)
				os.Exit(1)
			}

			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				os.Exit(1)
			}

			var gardens []models.Garden
			err = json.Unmarshal(body, &gardens)
			if err != nil {
				fmt.Println("Error unmarshalling response body:", err)
				os.Exit(1)
			}
			table := tablewriter.NewWriter(os.Stdout)

			table.SetHeader(
				[]string{"ID", "Name", "Location", "Description", "Created At", "Updated At"},
			)

			for _, v := range gardens {
				table.Append([]string{
					v.ID,
					v.Name,
					v.Location,
					v.Description,
					v.CreatedAt.Format(time.RFC822),
					v.UpdatedAt.Format(time.RFC822),
				})
			}
			table.Render()
		},
	}
}

func createGardenCmd(apiUrl string) *cobra.Command {
	createGardenCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new garden",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			location, _ := cmd.Flags().GetString("location")
			description, _ := cmd.Flags().GetString("description")

			garden := models.Garden{
				Name:        name,
				Location:    location,
				Description: description,
			}

			jsonData, err := json.Marshal(garden)
			if err != nil {
				fmt.Println("Error marshalling garden:", err)
				os.Exit(1)
			}

			response, err := http.Post(
				fmt.Sprintf("%s/gardens", apiUrl),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				fmt.Println("Error creating garden:", err)
				os.Exit(1)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				os.Exit(1)
			}

			if response.StatusCode != http.StatusCreated {
				fmt.Printf(
					"Error: Server returned status code %d: %s\n",
					response.StatusCode,
					string(body),
				)
				os.Exit(1)
			}

			fmt.Println("Garden created successfully!")
		},
	}
	createGardenCmd.Flags().StringP("name", "n", "", "Name of the garden")
	createGardenCmd.Flags().StringP("location", "l", "", "Location of the garden")
	createGardenCmd.Flags().StringP("description", "d", "", "Description of the garden")

	createGardenCmd.MarkFlagRequired("name")
	createGardenCmd.MarkFlagRequired("location")
	createGardenCmd.MarkFlagRequired("description")

	return createGardenCmd
}

func updateGardenCmd(apiUrl string) *cobra.Command {
	updateGardenCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a garden",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			name, _ := cmd.Flags().GetString("name")
			location, _ := cmd.Flags().GetString("location")
			description, _ := cmd.Flags().GetString("description")

			garden := models.Garden{
				Name:        name,
				Location:    location,
				Description: description,
			}

			jsonData, err := json.Marshal(garden)
			if err != nil {
				fmt.Println("Error marshalling garden:", err)
				os.Exit(1)
			}

			request, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("%s/gardens/%s", apiUrl, id),
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				fmt.Println("Error updating garden:", err)
				os.Exit(1)
			}

			request.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				fmt.Println("Error updating garden:", err)
				os.Exit(1)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				os.Exit(1)
			}

			if response.StatusCode != http.StatusNoContent {
				fmt.Printf(
					"Error: Server returned status code %d: %s\n",
					response.StatusCode,
					string(body),
				)
				os.Exit(1)
			}

			fmt.Println("Garden updated successfully!")
		},
	}

	updateGardenCmd.Flags().StringP("name", "n", "", "Name of the garden")
	updateGardenCmd.Flags().StringP("location", "l", "", "Location of the garden")
	updateGardenCmd.Flags().StringP("description", "d", "", "Description of the garden")

	return updateGardenCmd
}

func deleteGardenCmd(apiUrl string) *cobra.Command {
	deleteGardenCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a garden",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/gardens/%s", apiUrl, id), nil)
			if err != nil {
				fmt.Println("Error deleting garden:", err)
				os.Exit(1)
			}

			request.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				fmt.Println("Error deleting garden:", err)
				os.Exit(1)
			}

			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				os.Exit(1)
			}

			if response.StatusCode != http.StatusNoContent {
				fmt.Printf(
					"Error: Server returned status code %d: %s\n",
					response.StatusCode,
					string(body),
				)
				os.Exit(1)
			}

			fmt.Println("Garden deleted successfully!")
		},
	}

	return deleteGardenCmd
}
