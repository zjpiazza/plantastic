package cmd

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
	"github.com/zjpiazza/plantastic/pkg/models"
)

func tasksCmd(apiUrl string) *cobra.Command {
	tasksCmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage your garden tasks",
		Long:  `Create, update, and delete garden tasks`,
	}

	tasksCmd.AddCommand(listTasksCmd(apiUrl))
	tasksCmd.AddCommand(createTaskCmd(apiUrl))
	tasksCmd.AddCommand(updateTaskCmd(apiUrl))
	tasksCmd.AddCommand(deleteTaskCmd(apiUrl))

	return tasksCmd
}

func listTasksCmd(apiUrl string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Run: func(cmd *cobra.Command, args []string) {
			response, err := http.Get(fmt.Sprintf("%s/tasks", apiUrl))
			if err != nil {
				fmt.Println("Error getting tasks:", err)
				os.Exit(1)
			}

			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				os.Exit(1)
			}

			var gardens []models.Task
			err = json.Unmarshal(body, &gardens)
			if err != nil {
				fmt.Println("Error unmarshalling response body:", err)
				os.Exit(1)
			}
			table := tablewriter.NewWriter(os.Stdout)

			table.SetHeader(
				[]string{
					"ID",
					"Description",
					"Due Date",
					"Status",
					"Priority",
					"Created At",
					"Updated At",
				},
			)

			for _, v := range gardens {
				table.Append([]string{
					v.ID,
					v.Description,
					v.DueDate.Format(time.RFC822),
					v.Status,
					v.Priority,
					v.CreatedAt.Format(time.RFC822),
					v.UpdatedAt.Format(time.RFC822),
				})
			}
			table.Render()
		},
	}
}

func createTaskCmd(apiUrl string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new garden task",
		Run: func(cmd *cobra.Command, args []string) {
			description, _ := cmd.Flags().GetString("description")
			dueDateStr, _ := cmd.Flags().GetString("due-date")

			// Check that dueDate can be converted to a datetime object
			dueDate, err := time.Parse("01-02-2006", dueDateStr)
			if err != nil {
				fmt.Println("Due date is not properly formatted")
				os.Exit(1)
			}

			task := models.Task{
				Description: description,
				DueDate:     dueDate,
			}

			jsonData, err := json.Marshal(task)
			if err != nil {
				fmt.Println("Error marshalling task:")
			}

			response, err := http.Post(
				fmt.Sprintf("%s/tasks", apiUrl),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				fmt.Println("Error creating task:", err)
				os.Exit(1)
			}

			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("Error reading response: %d: %s\n", response.StatusCode, string(body))
				os.Exit(1)
			}
		},
	}
	cmd.Flag("description")
	cmd.Flag("due-date")

	cmd.MarkFlagRequired("description")
	cmd.MarkFlagRequired("due-date")
	return cmd
}

func updateTaskCmd(apiUrl string) *cobra.Command {
	updateGardenCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			description, _ := cmd.Flags().GetString("description")
			dueDateStr, _ := cmd.Flags().GetString("due-date")

			dueDate, err := time.Parse("01-02-2006", dueDateStr)
			if err != nil {
				fmt.Println("Invalid date format")
				os.Exit(1)
			}

			task := models.Task{
				Description: description,
				DueDate:     dueDate,
			}

			jsonData, err := json.Marshal(task)
			if err != nil {
				fmt.Println("Error marshalling task:", err)
				os.Exit(1)
			}

			request, err := http.NewRequest(
				"PUT",
				fmt.Sprintf("%s/tasks/%s", apiUrl, id),
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				fmt.Println("Error updating task:", err)
				os.Exit(1)
			}

			request.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				fmt.Println("Error updating task:", err)
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

			fmt.Println("Task updated successfully!")
		},
	}

	updateGardenCmd.Flags().StringP("description", "n", "", "Task description")
	updateGardenCmd.Flags().StringP("due-date", "l", "", "Date the task should be completed")

	return updateGardenCmd
}

func deleteTaskCmd(apiUrl string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a garden task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]

			request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/%s", apiUrl, id), nil)
			if err != nil {
				fmt.Println("Error deleting task:", err)
				os.Exit(1)
			}

			request.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				fmt.Println("Error deleting task:", err)
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
					"ErrorL Server returned status code %d: %s\n",
					response.StatusCode,
					string(body),
				)
				os.Exit(1)
			}

			fmt.Println("Task deleted successfully!")
		},
	}
}
