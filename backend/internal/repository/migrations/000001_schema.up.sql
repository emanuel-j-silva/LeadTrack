-- 000001_schema.up.sql
CREATE TABLE employee (
    id            SERIAL       PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(150) NOT NULL UNIQUE,
    position_name VARCHAR(100) NOT NULL
);

CREATE TABLE leader_lead (
    leader_id INT NOT NULL REFERENCES employee(id) ON DELETE CASCADE ON UPDATE CASCADE,
    lead_id   INT NOT NULL REFERENCES employee(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (leader_id, lead_id),
    CONSTRAINT chk_no_self_lead CHECK (leader_id <> lead_id)
);
CREATE INDEX idx_leader_lead_lead ON leader_lead(lead_id);

CREATE TABLE question (
    id     SERIAL       PRIMARY KEY,
    label  VARCHAR(120) NOT NULL UNIQUE,
    weight INT          NOT NULL CHECK (weight > 0)
);

CREATE TABLE evaluation (
    id           SERIAL PRIMARY KEY,
    evaluator_id INT NOT NULL REFERENCES employee(id),
    evaluated_id INT NOT NULL REFERENCES employee(id),
    week_start   DATE NOT NULL,
    score        NUMERIC(4,2) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (evaluator_id <> evaluated_id),
    UNIQUE (evaluator_id, evaluated_id, week_start)
);
CREATE INDEX idx_evaluation_evaluated ON evaluation(evaluated_id, created_at DESC);

CREATE TABLE evaluation_answer (
    evaluation_id INT NOT NULL REFERENCES evaluation(id),
    question_id   INT NOT NULL REFERENCES question(id),
    value         SMALLINT NOT NULL CHECK (value BETWEEN 1 AND 4),
    PRIMARY KEY (evaluation_id, question_id)
);
