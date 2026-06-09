-- Dummy sales database for Superset demo

CREATE TABLE IF NOT EXISTS sales (
    id          SERIAL PRIMARY KEY,
    order_date  DATE NOT NULL,
    region      VARCHAR(50) NOT NULL,
    product     VARCHAR(100) NOT NULL,
    category    VARCHAR(50) NOT NULL,
    quantity    INT NOT NULL,
    unit_price  NUMERIC(10,2) NOT NULL,
    revenue     NUMERIC(10,2) GENERATED ALWAYS AS (quantity * unit_price) STORED
);

DO $$
DECLARE
    pnames    TEXT[]    := ARRAY['Laptop Pro','Wireless Mouse','USB-C Hub','Desk Chair','Standing Desk','Notebook','Pen Set','Monitor 27"','Keyboard','Bookshelf'];
    pcats     TEXT[]    := ARRAY['Electronics','Electronics','Electronics','Furniture','Furniture','Stationery','Stationery','Electronics','Electronics','Furniture'];
    pprices   NUMERIC[] := ARRAY[1299.99, 29.99, 49.99, 399.99, 699.99, 4.99, 9.99, 449.99, 89.99, 249.99];
    regions   TEXT[]    := ARRAY['North', 'South', 'East', 'West'];
    idx       INT;
    ridx      INT;
    n         INT := 10;
BEGIN
    FOR i IN 1..2000 LOOP
        idx  := 1 + (floor(random() * n)::int % n);
        ridx := 1 + (floor(random() * 4)::int % 4);
        INSERT INTO sales (order_date, region, product, category, quantity, unit_price)
        VALUES (
            DATE '2024-01-01' + (random() * 365)::int,
            regions[ridx],
            pnames[idx],
            pcats[idx],
            (random() * 50 + 1)::int,
            pprices[idx]
        );
    END LOOP;
END $$;
