package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/zjpiazza/plantastic/pkg/models"
)

func bedsCmd(apiUrl string) *cobra.Command {
	bedsCmd := &cobra.Command{
		Use:   "beds",
		Short: "Manage garden beds",
		Long:  `Create, list, update and delete garden beds`,
	}

	bedsCmd.AddCommand(listBedsCmd(apiUrl))
	bedsCmd.AddCommand(createBedCmd(apiUrl))
	bedsCmd.AddCommand(updateBedCmd(apiUrl))
	bedsCmd.AddCommand(deleteBedCmd(apiUrl))

	return bedsCmd
}

func listBedsCmd(apiUrl string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all garden beds",
		Run: func(cmd *cobra.Command, args []string) {
			response, err := http.Get(fmt.Sprintf("%s/beds", apiUrl))
			if err != nil {
				fmt.Println("Error getting beds:", err)
				os.Exit(1)
			}

			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				os.Exit(1)
			}

			var beds []models.Bed
			err = json.Unmarshal(body, &beds)
			if err != nil {
				fmt.Println("Error unmarshalling response body:", err)
				os.Exit(1)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader(
				[]string{"ID", "Garden ID", "Name", "Type", "Size", "Soil Type", "Notes"},
			)

			for _, v := range beds {
				table.Append(
					[]string{v.ID, v.GardenID, v.Name, v.Type, v.Size, v.SoilType, v.Notes},
				)
			}
			table.Render()
		},
	}
}

func createBedCmd(apiUrl string) *cobra.Command {
	createBedCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new garden bed",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			bedType, _ := cmd.Flags().GetString("type")
			size, _ := cmd.Flags().GetString("size")
			soilType, _ := cmd.Flags().GetString("soil-type")
			notes, _ := cmd.Flags().GetString("notes")
			gardenID, _ := cmd.Flags().GetString("garden-id")

			bed := models.Bed{
				Name:     name,
				Type:     bedType,
				Size:     size,
				SoilType: soilType,
				Notes:    notes,
				GardenID: gardenID,
			}

			jsonData, err := json.Marshal(bed)
			if err != nil {
				fmt.Println("Error marshalling bed:", err)
				os.Exit(1)
			}

			response, err := http.Post(
				fmt.Sprintf("%s/beds", apiUrl),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				fmt.Println("Error creating bed:", err)
				os.Exit(1)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				os.Exit(1)
			}

			if response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusOK {
				fmt.Printf(
					"Error: Server returned status code %d: %s\n",
					response.StatusCode,
					string(body),
				)
				os.Exit(1)
			}

			fmt.Println("Garden bed created successfully!")

			// Try to display the created bed details if possible
			var createdBed models.Bed
			if err := json.Unmarshal(body, &createdBed); err == nil {
				fmt.Printf("Created bed: %s (ID: %s)\n", createdBed.Name, createdBed.ID)
			}
		},
	}
	createBedCmd.Flags().StringP("name", "n", "", "Name of the bed")
	createBedCmd.Flags().StringP("type", "t", "", "Type of the bed")
	createBedCmd.Flags().StringP("size", "s", "", "Size of the bed")
	createBedCmd.Flags().StringP("soil-type", "S", "", "Soil type of the bed")
	createBedCmd.Flags().StringP("notes", "N", "", "Notes of the bed")
	createBedCmd.Flags().StringP("garden-id", "g", "", "ID of the garden this bed belongs to")

	createBedCmd.MarkFlagRequired("name")
	createBedCmd.MarkFlagRequired("type")
	createBedCmd.MarkFlagRequired("size")
	createBedCmd.MarkFlagRequired("soil-type")
	createBedCmd.MarkFlagRequired("garden-id")

	return createBedCmd
}

func updateBedCmd(apiUrl string) *cobra.Command {
	updateBedCmd := &cobra.Command{
		Use:   "update <bed-id>",
		Short: "Update a garden bed",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			bedType, _ := cmd.Flags().GetString("type")
			size, _ := cmd.Flags().GetString("size")
			soilType, _ := cmd.Flags().GetString("soil-type")
			notes, _ := cmd.Flags().GetString("notes")
			gardenID, _ := cmd.Flags().GetString("garden-id")

			bed := models.Bed{
				Name:     name,
				Type:     bedType,
				Size:     size,
				SoilType: soilType,
				Notes:    notes,
				GardenID: gardenID,
			}

			jsonData, err := json.Marshal(bed)
			if err != nil {
				fmt.Println("Error marshalling bed:", err)
				os.Exit(1)
			}

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("%s/beds/%s", apiUrl, args[0]),
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				fmt.Println("Error creating request:", err)
				os.Exit(1)
			}

			req.Header.Set("Content-Type", "application/json")

			response, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("Error creating bed:", err)
				os.Exit(1)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				os.Exit(1)
			}

			if response.StatusCode != http.StatusOK {
				fmt.Printf(
					"Error: Server returned status code %d: %s\n",
					response.StatusCode,
					string(body),
				)
				os.Exit(1)
			}

			fmt.Println("Garden bed updated successfully!")

			// Try to display the updated bed details if possible
			var updatedBed models.Bed
			if err := json.Unmarshal(body, &updatedBed); err == nil {
				fmt.Printf("Updated bed: %s (ID: %s)\n", updatedBed.Name, updatedBed.ID)
			}
		},
	}
	updateBedCmd.Flags().StringP("name", "n", "", "Name of the bed")
	updateBedCmd.Flags().StringP("type", "t", "", "Type of the bed")
	updateBedCmd.Flags().StringP("size", "s", "", "Size of the bed")
	updateBedCmd.Flags().StringP("soil-type", "S", "", "Soil type of the bed")
	updateBedCmd.Flags().StringP("notes", "N", "", "Notes of the bed")
	updateBedCmd.Flags().StringP("garden-id", "g", "", "ID of the garden this bed belongs to")

	return updateBedCmd
}

func deleteBedCmd(apiUrl string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a garden bed",
		Run: func(cmd *cobra.Command, args []string) {
			req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/beds/%s", apiUrl, args[0]), nil)
			if err != nil {
				fmt.Println("Error deleting bed:", err)
				os.Exit(1)
			}
			req.Header.Set("Content-Type", "application/json")

			response, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("Error deleting bed:", err)
				os.Exit(1)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				fmt.Printf("Error: Server returned status code %d\n", response.StatusCode)
				os.Exit(1)
			}

			fmt.Println("Garden bed deleted successfully!")
		},
	}
}
