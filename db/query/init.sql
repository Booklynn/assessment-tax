CREATE TABLE "allowances" (
    "id" bigserial PRIMARY KEY,
    "donation" bigint NOT NULL,
    "personal" bigint NOT NULL,
    "k-receipt" bigint NOT NULL
);

CREATE INDEX ON "allowances" ("donation", "personal", "k-receipt");

COMMENT ON COLUMN "allowances"."personal" IS 'mininum is 10000 and cannot be greater than 100000';

COMMENT ON COLUMN "allowances"."k-receipt" IS 'mininum is 0 and cannot be greater than 100000';

INSERT INTO "allowances" (
    "donation", "personal", "k-receipt"
) VALUES (
    100000, 60000, 50000
);
