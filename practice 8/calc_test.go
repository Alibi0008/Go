package main

import "testing"

func TestAdd(t *testing.T) {
	got := Add(2, 3)
	want := 5

	if got != want {
		t.Errorf("Add(2, 3) = %d; want %d", got, want)
	}
}

func TestAddTableDriven(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{name: "both positive", a: 2, b: 3, want: 5},
		{name: "positive plus zero", a: 5, b: 0, want: 5},
		{name: "negative plus positive", a: -1, b: 4, want: 3},
		{name: "both negative", a: -2, b: -3, want: -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSubtractTableDriven(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{name: "both positive numbers", a: 9, b: 4, want: 5},
		{name: "positive minus zero", a: 7, b: 0, want: 7},
		{name: "negative minus positive", a: -3, b: 4, want: -7},
		{name: "both negative", a: -5, b: -2, want: -3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Subtract(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Subtract(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := Divide(20, 5)
		if err != nil {
			t.Fatalf("Divide(20, 5) returned unexpected error: %v", err)
		}

		if got != 4 {
			t.Errorf("Divide(20, 5) = %d; want %d", got, 4)
		}
	})

	t.Run("division by zero", func(t *testing.T) {
		_, err := Divide(20, 0)
		if err == nil {
			t.Fatal("Divide(20, 0) expected an error, got nil")
		}

		if err.Error() != "division by zero" {
			t.Errorf("Divide(20, 0) error = %q; want %q", err.Error(), "division by zero")
		}
	})
}
