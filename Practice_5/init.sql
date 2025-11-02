CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    genre TEXT NOT NULL,
    price INT NOT NULL
);

INSERT INTO books (title, author, genre, price) VALUES
('War and Peace', 'Leo Tolstoy', 'fiction', 500),
('Crime and Punishment', 'Fyodor Dostoevsky', 'fiction', 450),
('The Hobbit', 'J.R.R. Tolkien', 'fantasy', 350),
('The Martian', 'Andy Weir', 'sci-fi', 400),
('Sapiens', 'Yuval Harari', 'non-fiction', 600);
