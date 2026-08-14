-- name: HydrateCoaster :one
SELECT sqlc.embed(coasters)
FROM coasters
WHERE coasters.id = @id;

-- Candidates for (re-)embedding.. either never embedded, or embedded under a different
-- model than the parameterized one.
--
-- Park is embedded so that the doc producer as as much metadata abt the coaster as possible to
-- improve embedding outcomes.
--
-- name: ListCoastersNeedingEmbedding :many
SELECT sqlc.embed(coasters), sqlc.embed(parks)
FROM coasters
JOIN parks ON parks.id = coasters.park_id
WHERE coasters.profile_embedding IS NULL
   OR coasters.profile_embedding_model IS DISTINCT FROM @model::text
ORDER BY coasters.id
LIMIT @row_limit;

-- Nearest neighbours to an embedded query vector, closest first.
-- <=> is cosine distance. non-embedded coasters are excluded.
-- name: SearchCoasters :many
SELECT sqlc.embed(coasters)
FROM coasters
WHERE profile_embedding IS NOT NULL
ORDER BY profile_embedding <=> @query_embedding
LIMIT @row_limit;

-- Stores the vector alongside the exact text and model that produced it, so a
-- later run can tell whether a row is stale. updated_at is deliberately left
-- alone: the coaster itself did not change, only its derived embedding.
-- name: SetCoasterEmbedding :exec
UPDATE coasters
SET search_profile          = @search_profile::text,
    profile_embedding       = @profile_embedding,
    profile_embedding_model = @profile_embedding_model::text,
    profile_embedded_at     = now()
WHERE id = @id;
