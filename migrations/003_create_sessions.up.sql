CREATE TABLE sessions (
    id             SERIAL PRIMARY KEY,
    campaign_id    INT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    session_number INT NOT NULL,
    title          TEXT,
    summary        TEXT,
    played_on      DATE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(campaign_id, session_number)
);

CREATE TABLE session_npcs (
    session_id INT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    npc_id     INT NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    introduced BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (session_id, npc_id)
);

CREATE TABLE session_locations (
    session_id  INT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    location_id INT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    PRIMARY KEY (session_id, location_id)
);

CREATE TABLE session_items (
    session_id INT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    item_id    INT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    PRIMARY KEY (session_id, item_id)
);
