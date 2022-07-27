BEGIN
;

CREATE TABLE IF NOT EXISTS "libraries" (
    "id" SERIAL,
    "create_at" TIMESTAMP WITH TIME ZONE NOT NULL,
    "name" VARCHAR(250) NOT NULL,
    "path" VARCHAR(250) NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT library_id_unique UNIQUE ("id"),
    CONSTRAINT library_path_unique UNIQUE ("path")
);

CREATE TABLE IF NOT EXISTS "books" (
    "id" SERIAL,
    "library_id" INT NOT NULL,
    "create_at" TIMESTAMP WITH TIME ZONE NOT NULL,
    "update_at" TIMESTAMP WITH TIME ZONE NOT NULL,
    "path" VARCHAR(250) NOT NULL,
    "name" VARCHAR(250) NOT NULL,
    "writer" VARCHAR(250) NOT NULL,
    "volume" INTEGER NOT NULL,
    "summary" text NOT NULL,
    "files" text NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT books_id_unique UNIQUE ("id"),
    CONSTRAINT books_path_unique UNIQUE ("path")
);

CREATE INDEX IF NOT EXISTS "books_libid_index" ON "books" ("library_id");
CREATE INDEX IF NOT EXISTS "books_name_index" ON "books" ("name");
CREATE INDEX IF NOT EXISTS "books_writer_index" ON "books" ("writer");
CREATE INDEX IF NOT EXISTS "books_path_index" ON "books" ("path");

COMMIT;