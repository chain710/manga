BEGIN;

CREATE INDEX IF NOT EXISTS "books_name_fts" ON "books" using gin(to_tsvector('jiebaqry', name));

CREATE INDEX IF NOT EXISTS "books_writer_fts" ON "books" using gin(to_tsvector('jiebaqry', writer));

END;
