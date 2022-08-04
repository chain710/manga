BEGIN;
CREATE TABLE IF NOT EXISTS "libraries" (
  "id" SERIAL,
  "create_at" TIMESTAMP WITH TIME ZONE NOT NULL,
  "scan_at" TIMESTAMP WITH TIME ZONE NOT NULL,
  "name" VARCHAR(250) NOT NULL,
  "path" VARCHAR(250) NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT library_id_unique UNIQUE ("id"),
  CONSTRAINT library_path_unique UNIQUE ("path")
);
CREATE TABLE IF NOT EXISTS "books" (
  "id" SERIAL,
  "library_id" INT NOT NULL REFERENCES "libraries" ("id") ON DELETE CASCADE,
  "create_at" TIMESTAMP WITH TIME ZONE NOT NULL,
  "update_at" TIMESTAMP WITH TIME ZONE NOT NULL,
  "path_mod_at" TIMESTAMP WITH TIME ZONE NOT NULL,
  "path" VARCHAR(250) NOT NULL,
  "name" VARCHAR(250) NOT NULL,
  "writer" VARCHAR(250) NOT NULL,
  "volume" INTEGER NOT NULL,
  "summary" TEXT NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT books_id_unique UNIQUE ("id"),
  CONSTRAINT books_path_unique UNIQUE ("path")
);

CREATE INDEX IF NOT EXISTS "books_libid_index" ON "books" ("library_id");
CREATE INDEX IF NOT EXISTS "books_name_index" ON "books" ("name");
CREATE INDEX IF NOT EXISTS "books_writer_index" ON "books" ("writer");
CREATE INDEX IF NOT EXISTS "books_path_index" ON "books" ("path");

CREATE TABLE IF NOT EXISTS "volumes" (
  "id" SERIAL,
  "book_id" INT NOT NULL REFERENCES "books" ("id") ON DELETE CASCADE,
  "create_at" TIMESTAMP WITH TIME ZONE NOT NULL,
  "path" VARCHAR(250) NOT NULL,
  "title" VARCHAR(250) NOT NULL, /* title of this volume */
  "volume" int NOT NULL, /* number of volume, for sort; extra volume always equal 0 */
  "page_count" int NOT NULL, /* how many pages(files) */
  "files" TEXT NOT NULL, /* files in archive, should be sorted */
  PRIMARY KEY ("id"),
  CONSTRAINT book_volumes_bid_vol_unique UNIQUE ("book_id", "volume"),
  CONSTRAINT book_volumes_path_unique UNIQUE ("path") /* for now, path should be unique */
);

CREATE INDEX IF NOT EXISTS "volumes_book_id_index" ON "volumes" ("book_id");

CREATE TABLE IF NOT EXISTS "volume_progress" (
    "create_at" TIMESTAMP WITH TIME ZONE NOT NULL,
    "update_at" TIMESTAMP WITH TIME ZONE NOT NULL,
    "book_id" INT NOT NULL REFERENCES "books" ("id"),
    "volume_id" INT NOT NULL REFERENCES "volumes" ("id") ON DELETE CASCADE,
    "complete" BOOLEAN NOT NULL, /* for filtering */
    "page" INT NOT NULL,
    PRIMARY KEY ("volume_id") /* user id */
);

CREATE INDEX IF NOT EXISTS "volume_progress_bid_index" ON "volume_progress" ("book_id");

CREATE TABLE IF NOT EXISTS "volume_thumbnail" (
    "id" INT NOT NULL REFERENCES "volumes" ("id") ON DELETE CASCADE,
    "thumbnail" BYTEA NOT NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS "book_thumbnail" (
    "id" INT NOT NULL REFERENCES "books" ("id") ON DELETE CASCADE,
     "thumbnail" BYTEA NOT NULL,
     PRIMARY KEY ("id")
);

COMMIT;