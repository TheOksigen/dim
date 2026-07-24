CREATE TABLE IF NOT EXISTS exam_results (
  fin_code varchar(7) PRIMARY KEY CHECK (fin_code ~ '^[A-Z0-9]{7}$'),
  first_name varchar(64) NOT NULL,
  last_name varchar(64) NOT NULL,
  father_name varchar(64) NOT NULL,
  exam_date date NOT NULL,

  subject_1_name varchar(48) NOT NULL,
  subject_1_score smallint NOT NULL CHECK (subject_1_score BETWEEN 0 AND 100),
  subject_1_answers varchar(128) NOT NULL,
  subject_2_name varchar(48) NOT NULL,
  subject_2_score smallint NOT NULL CHECK (subject_2_score BETWEEN 0 AND 100),
  subject_2_answers varchar(128) NOT NULL,
  subject_3_name varchar(48) NOT NULL,
  subject_3_score smallint NOT NULL CHECK (subject_3_score BETWEEN 0 AND 100),
  subject_3_answers varchar(128) NOT NULL,

  total_score smallint NOT NULL CHECK (total_score BETWEEN 0 AND 300),
  passing_status boolean NOT NULL,
  result_version integer NOT NULL DEFAULT 1 CHECK (result_version > 0),
  updated_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE exam_results IS 'Synthetic examination records. FIN primary-key B-tree is the lookup index.';

