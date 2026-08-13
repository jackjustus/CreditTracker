-- Deterministic development seed. Safe to re-run: fixed ids and timestamps
-- mean every statement is a no-op after the first apply.

INSERT INTO users (id, name) VALUES
    ('00000000-0000-0000-0000-000000000001', 'localuser')
ON CONFLICT (id) DO NOTHING;

-- INSERT INTO parks (id, name, city, country) VALUES
--     ('019fca53-0b17-7862-9c0a-706d42fb5aab',
--      'Disneyland',
--      'Anaheim',
--      'United States')
-- ON CONFLICT (id) DO NOTHING;
--
-- INSERT INTO coasters (id, park_id, name, manufactured_at) VALUES
--     ('019fca53-a87a-7be7-bdbf-62e36af8b28a',
--      '019fca53-0b17-7862-9c0a-706d42fb5aab',
--      'Matterhorn',
--      '1959-06-14T00:00:00Z')
-- ON CONFLICT (id) DO NOTHING;

-- INSERT INTO ride_events (user_id, coaster_id, created_at) VALUES
--     ('00000000-0000-0000-0000-000000000001',
--      '019fca53-a87a-7be7-bdbf-62e36af8b28a',
--      '2026-01-01T12:00:00Z')
-- ON CONFLICT (user_id, coaster_id, created_at) DO NOTHING;
