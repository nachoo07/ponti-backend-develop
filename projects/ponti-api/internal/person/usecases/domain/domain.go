package domain

type Person struct {
	ID         string   // Identificador único.
	FirstName  string   // Nombre.
	LastName   string   // Apellido.
	Age        int      // Edad.
	Gender     string   // Género.
	NationalID int64    // Identificador Nacional (por ejemplo, DNI).
	Phone      string   // Teléfono.
	Deleted    bool     // Indica si fue eliminada (soft delete).
	Interests  []string // Áreas de interés.
	Hobbies    []string // Hobbies.
}
