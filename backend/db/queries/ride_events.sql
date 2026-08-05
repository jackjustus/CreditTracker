
-- name: GetRideEvents :many
SELECT sqlc.embed(ride_events), sqlc.embed(coasters)
FROM ride_events
JOIN coasters ON ride_events.coaster_id = coasters.id
ORDER BY ride_events.created_at DESC;


-- name: CreateRideEvent :exec
INSERT INTO ride_events (coaster_id, created_at)
VALUES (
        @coaster_id,
        @created_at
       );
