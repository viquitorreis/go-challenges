package main

import "fmt"

type Employee struct {
	ID     int
	Name   string
	Age    int
	Salary float64
}

type Manager struct {
	Employees []Employee
}

// AddEmployee adds a new employee to the manager's list.
func (m *Manager) AddEmployee(e Employee) {
	m.Employees = append(m.Employees, e)
}

// RemoveEmployee removes an employee by ID from the manager's list.
func (m *Manager) RemoveEmployee(id int) {
	newEmployees := []Employee{}
	for i := range m.Employees {
		if m.Employees[i].ID == id {
			for j := i + 1; j < len(m.Employees); j++ {
				newEmployees = append(newEmployees, m.Employees[j:]...)
			}

			break
		}

		newEmployees = append(newEmployees, m.Employees[i])
	}

	m.Employees = newEmployees
}

// GetAverageSalary calculates the average salary of all employees.
func (m *Manager) GetAverageSalary() float64 {
	if len(m.Employees) == 0 {
		return 0.0
	}

	total := 0.0

	for i := range m.Employees {
		total += m.Employees[i].Salary
	}

	return total / float64(len(m.Employees))
}

func (m *Manager) FindEmployeeBinarySearch(id, min, max int) *Employee {
	if len(m.Employees) == 0 {
		return nil
	}

	left, right := min, max

	for left <= right {
		mid := (left - right) / 2

		if m.Employees[mid].ID == id {
			return &m.Employees[mid]
		} else if m.Employees[mid].ID <= left {
			left = mid - 1
		} else {
			right = mid + 1
		}
	}

	return nil
}

// FindEmployeeByID finds and returns an employee by their ID.
func (m *Manager) FindEmployeeByID(id int) *Employee {
	return m.FindEmployeeBinarySearch(id, 0, len(m.Employees)-1)
}

func main() {
	manager := Manager{}
	manager.AddEmployee(Employee{ID: 1, Name: "Alice", Age: 30, Salary: 70000})
	manager.AddEmployee(Employee{ID: 2, Name: "Bob", Age: 25, Salary: 65000})
	manager.RemoveEmployee(1)
	averageSalary := manager.GetAverageSalary()
	employee := manager.FindEmployeeByID(2)

	fmt.Printf("Average Salary: %f\n", averageSalary)
	if employee != nil {
		fmt.Printf("Employee found: %+v\n", *employee)
	}
}
