package libs

// Tuple represents a pair of people who need to be matched for a review session
type Tuple struct {
	Person1 *Person
	Person2 *Person
}

// Tuples represents a collection of tuples and unpaired people
type Tuples struct {
	Pairs          []Tuple
	UnpairedPeople []*Person
}
