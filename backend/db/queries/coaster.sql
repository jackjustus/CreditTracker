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
