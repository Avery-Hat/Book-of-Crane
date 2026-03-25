CREATE TABLE npc_factions (
    npc_id     INT NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    faction_id INT NOT NULL REFERENCES factions(id) ON DELETE CASCADE,
    role       TEXT,
    PRIMARY KEY (npc_id, faction_id)
);

CREATE TABLE npc_locations (
    npc_id      INT NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    location_id INT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    context     TEXT,
    PRIMARY KEY (npc_id, location_id)
);

CREATE TABLE faction_locations (
    faction_id   INT NOT NULL REFERENCES factions(id) ON DELETE CASCADE,
    location_id  INT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    relationship TEXT,
    PRIMARY KEY (faction_id, location_id)
);

CREATE TABLE npc_relationships (
    npc_id_1     INT NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    npc_id_2     INT NOT NULL REFERENCES npcs(id) ON DELETE CASCADE,
    relationship TEXT NOT NULL,
    notes        TEXT,
    PRIMARY KEY (npc_id_1, npc_id_2),
    CHECK (npc_id_1 < npc_id_2)
);

-- Full-text search indexes
CREATE INDEX idx_npcs_search ON npcs
    USING GIN (to_tsvector('english', coalesce(name,'') || ' ' || coalesce(description,'') || ' ' || coalesce(notes,'')));

CREATE INDEX idx_locations_search ON locations
    USING GIN (to_tsvector('english', coalesce(name,'') || ' ' || coalesce(description,'') || ' ' || coalesce(notes,'')));

CREATE INDEX idx_factions_search ON factions
    USING GIN (to_tsvector('english', coalesce(name,'') || ' ' || coalesce(description,'') || ' ' || coalesce(notes,'')));

CREATE INDEX idx_sessions_search ON sessions
    USING GIN (to_tsvector('english', coalesce(title,'') || ' ' || coalesce(summary,'')));
