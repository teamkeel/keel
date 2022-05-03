package testdata

const ReferenceExample = `
model Book {
	fields {
	  title Text
	  isbn Text {
		@unique
	  }
	  authors Author[]
	}
	functions {
	  create createBook(title, authors)
	  get book(id)
	  get bookByIsbn(isbn)
	}
  }`