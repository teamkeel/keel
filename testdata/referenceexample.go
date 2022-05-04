package testdata

const ReferenceExample = `
model Book {
	fields {
	  title Text
	  isbn Text {
		@unique
	  }
	  // Not valid
	  // authors Author[]
	}
	functions {
	  create createBook(title, authors)
	  get book(id)
	  get bookByIsbn(isbn)
	}
  }`
