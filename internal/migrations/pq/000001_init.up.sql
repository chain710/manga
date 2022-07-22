CREATE TABLE IF NOT EXISTS books (
    "id" SERIAL,
    "create_at" TIMESTAMP WITH TIME ZONE NOT NULL,
    "update_at" TIMESTAMP WITH TIME ZONE NOT NULL,
    "delete_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    "name" VARCHAR(250) NOT NULL,
    "writer" VARCHAR(250) NOT NULL,
    "volume" INTEGER NOT NULL,
    "page_count" BIGINT NOT NULL,
    "summary" text NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT books_id_unique UNIQUE ("id")
);