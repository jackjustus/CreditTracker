-- +goose Up
CREATE TABLE parks (
    id              uuid NOT NULL   PRIMARY KEY DEFAULT uuidv7(),
    name            text NOT NULL,
    city            text NOT NULL,
    country         text NOT NULL,

    -- external_id and source refer to the place which this row was sourced from.
    external_id     text, --ex: wikidata uuid of a park
    external_source text, --ex: "wikidata"

    created_at      timestamptz     NOT NULL    DEFAULT now(),
    updated_at      timestamptz     NOT NULL    DEFAULT now(),

    UNIQUE (name, city, country),
    CONSTRAINT external_ref
        CHECK ((external_source IS NULL) = (external_id IS NULL)),
    UNIQUE (external_id, external_source)

);
CREATE TABLE coasters (
    id              uuid        NOT NULL    PRIMARY KEY DEFAULT uuidv7(),
    park_id         uuid        NOT NULL    REFERENCES parks (id),
    name            text        NOT NULL,
    manufactured_at timestamptz,

    -- external_id and source refer to the place which this row was sourced from.
    external_id     text, --ex: wikidata uuid of a coaster
    external_source text, --ex: "wikidata"

    created_at      timestamptz     NOT NULL    DEFAULT now(),
    updated_at      timestamptz     NOT NULL    DEFAULT now(),

    CONSTRAINT external_ref
        CHECK ((external_source IS NULL) = (external_id IS NULL)),
    UNIQUE (external_id, external_source)
);
CREATE TABLE ride_events (
    coaster_id      uuid        NOT NULL    REFERENCES coasters (id),
    created_at      timestamptz NOT NULL    DEFAULT now(),

    PRIMARY KEY (coaster_id, created_at)
);




-- +goose Down
DROP TABLE ride_events;
DROP TABLE coasters;
DROP TABLE parks;
