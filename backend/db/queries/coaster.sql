-- name: HydrateCoaster :one
SELECT sqlc.embed(coasters)
FROM coasters
WHERE coasters.id = @id;
