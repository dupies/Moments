import * as Knex from "knex";

export async function up(knex: Knex): Promise<void> {
  return knex.schema.createTable("mmnt_items", async (table) => {
    table.integer("id").primary();
    table.integer("type_id");
    table.text("owner_address");
    table.timestamps(true, true);
  });
}

export async function down(knex: Knex): Promise<void> {
  await knex.raw(
    "ALTER TABLE mmnt_items DROP CONSTRAINT mmnt_items_pkey CASCADE"
  );
  return knex.schema.dropTable("mmnt_items");
}
