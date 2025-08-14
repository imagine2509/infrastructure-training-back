-- Create notes table
CREATE TABLE IF NOT EXISTS notes
(
    id         SERIAL PRIMARY KEY,
    text       TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on created_at for better performance when ordering
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes (created_at DESC);

-- Create function to automatically update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at on UPDATE
CREATE OR REPLACE TRIGGER update_notes_updated_at
    BEFORE UPDATE
    ON notes
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();