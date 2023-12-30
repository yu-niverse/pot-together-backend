package query

import (
	"database/sql"
	"pottogether/pkg/mariadb"
)

type Ingredient struct {
	ID          int    `json:"ingredientID"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Interval    int    `json:"interval"`
	Requirement string `json:"requirement"`
}

// get all ingredients
func GetIngredients() ([]Ingredient, error) {
	query := `
		SELECT id, name, image, time_interval, requirement
		FROM ingredient
	`
	rows, err := mariadb.DB.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return []Ingredient{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	var ingredients []Ingredient
	for rows.Next() {
		var ingredient Ingredient
		err = rows.Scan(&ingredient.ID, &ingredient.Name, &ingredient.Image, &ingredient.Interval, &ingredient.Requirement)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ingredient)
	}
	return ingredients, nil
}

func AddIngredient(ingredient Ingredient) (int, error) {
	query := `
		INSERT INTO ingredient (name, image, time_interval, requirement)
		VALUES (?, ?, ?, ?)
	`
	result, err := mariadb.DB.Exec(query, ingredient.Name, ingredient.Image, ingredient.Interval, ingredient.Requirement)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil
}
