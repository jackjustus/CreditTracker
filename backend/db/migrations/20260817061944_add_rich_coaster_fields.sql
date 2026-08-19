-- +goose Up

CREATE TABLE manufacturers (
                               manufacturer_id uuid    NOT NULL    PRIMARY KEY  DEFAULT uuidv7(),
                               name            text    NOT NULL,
                               city            text,
                               state           text,
                               country         text    NOT NULL
);

CREATE TABLE coaster_styles (
    coaster_style_id        uuid        NOT NULL        PRIMARY KEY     DEFAULT uuidv7(),
    manufacturer_id         uuid        NOT NULL        REFERENCES manufacturers (manufacturer_id),
    material                VARCHAR(16) NOT NULL,       -- wooden, hybrid, or steel.
    seating_style           text        NOT NULL,       -- sitdown, flying, inverted, ...

    -- used when a richer description is available than the stricter fields above.
    -- consumers should prefer this field to generate a desc over the above fields if it is populated.
    -- EX:
    --  (gerstlauer) infinity dive [w/ desc] vs. steel dive [w/o].
    --  (vekoma) next-gen flying   [w/ desc] vs. steel flying [w/o].
    description             text
);

ALTER TABLE coasters
    ADD COLUMN height           int,
    ADD COLUMN length           int,
    ADD COLUMN top_speed_mph    float,
    ADD COLUMN inversion_count  int,
    -- TODO: make manufacturer_id NOT NULL once backfilled.
    ADD COLUMN coaster_style_id uuid    REFERENCES coaster_styles (coaster_style_id),
    ADD COLUMN opening_date     date,
    ADD COLUMN closing_date     date
;


ALTER TABLE parks
    ADD COLUMN state VARCHAR(32);


-- +goose Down

DROP TABLE manufacturers;
DROP TABLE coaster_styles;
ALTER TABLE coasters
    DROP COLUMN height           ,
    DROP COLUMN length           ,
    DROP COLUMN top_speed_mph    ,
    DROP COLUMN inversion_count  ,
    DROP COLUMN coaster_style_id ,
    DROP COLUMN opening_date     ,
    DROP COLUMN closing_date;