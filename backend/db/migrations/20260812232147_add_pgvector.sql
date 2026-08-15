-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE coasters
    ADD COLUMN search_profile          text,        -- text passed to embedding model
    ADD COLUMN profile_embedding       vector(768),
    ADD COLUMN profile_embedding_model text,        -- text literal of embedding model used, ex: nomic-embed-text
    ADD COLUMN profile_embedded_at     timestamptz;

-- +goose Down
ALTER TABLE coasters
    DROP COLUMN search_profile          ,
    DROP COLUMN profile_embedding       ,
    DROP COLUMN profile_embedding_model ,
    DROP COLUMN profile_embedded_at     ;

DROP EXTENSION IF EXISTS vector;