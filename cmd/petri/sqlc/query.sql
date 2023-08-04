-- name: FindUserByEmail :one
SELECT "id", "email", "role", "createdAt", "updatedAt" FROM "User" WHERE ("email" = $1 AND 1=1) LIMIT $2 OFFSET $3;
-- name: FindUserByID :one
SELECT "id", "email", "role", "createdAt", "updatedAt" FROM "User" WHERE ("id" = $1 AND 1=1) LIMIT $2 OFFSET $3;
-- name: ListNetsByAuthor :many
SELECT "id", "name", "description", "initialMarking", "createdAt", "updatedAt", "authorID", "parentID", "deviceId" FROM "Net" WHERE "authorID" = $1 OFFSET $2;
-- name: FindNetByID :one
SELECT "id", "name", "description", "initialMarking", "deviceId" FROM "Net" WHERE ("id" = $1 AND 1=1) LIMIT $2 OFFSET $3;
-- name: FindPlacesByNetID :many
SELECT "id", "name", "bound" FROM "Place" WHERE ("id") IN (SELECT "t0"."B" FROM "_NetToPlace" AS "t0" INNER JOIN "Net" AS "j0" ON ("j0"."id") = ("t0"."A") WHERE ("j0"."id" = $1 AND "t0"."B" IS NOT NULL)) OFFSET $2;
-- name: FindTransitionsByNetID :many
SELECT "id", "name" FROM "Transition" WHERE ("id") IN (SELECT "t0"."B" FROM "_NetToTransition" AS "t0" INNER JOIN "Net" AS "j0" ON ("j0"."id") = ("t0"."A") WHERE ("j0"."id" = $1 AND "t0"."B" IS NOT NULL)) OFFSET $2;
-- name: FindArcsByNetID :many
SELECT "id", "fromPlace", "placeID", "transitionID" FROM "Arc" WHERE "netID" = $1 OFFSET $2;
-- name: FindNetsByParentID :many
SELECT "id", "name", "description", "initialMarking", "deviceId" FROM "Net" WHERE "parentID" = $1 OFFSET $2;
-- name: FindDevicesByUser :many
SELECT "id", "name", "description" FROM "Device" WHERE "authorID" = $1 OFFSET $2;
-- name: FindNetsByDeviceID :many
SELECT "id", "name", "description", "initialMarking", "deviceId" FROM "Net" WHERE "deviceId" = $1 OFFSET $2;
-- name: FindInstancesByUserID :many
SELECT "id", "name", "addr" FROM "DeviceInstance" WHERE "authorID" = $1 OFFSET $2;
-- name: FindInstancesByDeviceID :many
SELECT "id", "name", "addr" FROM "DeviceInstance" WHERE "deviceId" = $1 OFFSET $2;
-- name: FindInstancesByAddr :many
SELECT "id", "name", "addr" FROM "DeviceInstance" WHERE "addr" = $1 OFFSET $2;
-- name: FindTransitionsWithNonNullEventsByNetID :many
SELECT "id", "name" FROM "Transition" WHERE (("id") IN (SELECT "t0"."B" FROM "_NetToTransition" AS "t0" INNER JOIN "Net" AS "j0" ON ("j0"."id") = ("t0"."A") WHERE ("j0"."id" = $1 AND "t0"."B" IS NOT NULL)) AND ("id") IN (SELECT "t0"."B" FROM "_EventToTransition" AS "t0" INNER JOIN "Event" AS "j0" ON ("j0"."id") = ("t0"."A") WHERE (1=1 AND "t0"."B" IS NOT NULL))) OFFSET $2;
-- name: FindEventsByTransitionID :many
SELECT "id", "name" FROM "Event" WHERE ("id") IN (SELECT "t0"."A" FROM "_EventToTransition" AS "t0" INNER JOIN "Transition" AS "j0" ON ("j0"."id") = ("t0"."B") WHERE ("j0"."id" = $1 AND "t0"."A" IS NOT NULL)) OFFSET $2;
-- name: FindFieldsByEvent :many
SELECT "id", "name", "type", "condition", "eventID", "createdAt", "updatedAt" FROM "Field" WHERE "eventID" = $1 OFFSET $2;
