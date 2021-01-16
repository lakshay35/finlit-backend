CREATE TABLE IF NOT EXISTS users (
  user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  first_name VARCHAR (255) NOT NULL,
	 last_name VARCHAR (255) NOT NULL,
	 email VARCHAR (255) UNIQUE NOT NULL,
	 phone VARCHAR (255) NOT NULL,
   google_id VARCHAR UNIQUE NOT NULL,
  registration_date TIMESTAMP default current_timestamp
);

CREATE TABLE IF NOT EXISTS roles (
  role_id SERIAL PRIMARY KEY,
  role_name VARCHAR (50) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_objects (
  audit_object_id SERIAL PRIMARY KEY,
  audit_object_name VARCHAR (255) NOT NULL,
  foreign_id VARCHAR (255) NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_types (
  audit_type_id SERIAL PRIMARY KEY,
  audit_type_name VARCHAR (255) NOT NULL
);

CREATE TABLE IF NOT EXISTS external_accounts (
  external_account_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  institutional_id VARCHAR (255) NOT NULL,
  user_id UUID NOT NULL,
  access_token VARCHAR (255) NOT NULL,
  account_name VARCHAR NOT NULL DEFAULT '',
  FOREIGN KEY (user_id)
      REFERENCES users (user_id)
);

CREATE TABLE IF NOT EXISTS budgets (
  budget_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  budget_name VARCHAR (255) NOT NULL,
  owner_id UUID NOT NULL,
	FOREIGN KEY (owner_id)
      REFERENCES users (user_id)
);

CREATE TABLE IF NOT EXISTS budgets_new (
  budget_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  budget_name VARCHAR (255) NOT NULL DEFAULT '',
  owner_id UUID NOT NULL,
	FOREIGN KEY (owner_id)
      REFERENCES users (user_id)
);

CREATE TABLE IF NOT EXISTS expense_charge_cycles (
  expense_charge_cycle_id SERIAL PRIMARY KEY,
  unit VARCHAR (50) NOT NULL
);

CREATE TABLE IF NOT EXISTS	 expenses (
  expense_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  budget_id UUID NOT NULL,
  expense_name VARCHAR (255),
  expense_value NUMERIC NOT NULL,
  expense_description VARCHAR,
  expense_charge_cycle_id INT NOT NULL,
  FOREIGN KEY (budget_id)
      REFERENCES budgets (budget_id),
  FOREIGN KEY (expense_charge_cycle_id)
      REFERENCES expense_charge_cycles (expense_charge_cycle_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
  user_role_id SERIAL PRIMARY KEY,
  user_id UUID NOT NULL,
  role_id INT NOT NULL,
  budget_id UUID NOT NULL,
  FOREIGN KEY (budget_id)
      REFERENCES budgets (budget_id),
  FOREIGN KEY (user_id)
      REFERENCES users (user_id),
  FOREIGN KEY (role_id)
      REFERENCES roles (role_id)
);

