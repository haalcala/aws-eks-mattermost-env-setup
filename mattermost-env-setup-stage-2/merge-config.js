const mysql = require("mysql");
const { program } = require("commander");
const fs = require("fs");

program
	.option("-b, --base-config <base-config>", "Base config", "config/config.json")
	.option("-m, --merged-config <merged-config>", "Merged config", "merged-config.json")
	.option("-h, --db-host <db-host>", "DB Host", "localhost")
	.option("-P, --db-port <db-port>", "DB Port", 3306)
	.option("-u, --db-user <db-user>", "DB User", "root")
	.option("-p, --db-pass <db-pass>", "DB Pass", "changeme")
	.option("-n, --db-name <db-name>", "DB Name", "test");

program.parse(process.argv);

console.log("program:", program);

let rawdata = fs.readFileSync(program.baseConfig);
let new_config = JSON.parse(rawdata);
console.log("----------------------", new_config);

const connection = mysql.createConnection({
	host: program.dbHost,
	port: program.dbPort,
	user: program.dbUser,
	password: program.dbPass,
	database: program.dbName,
});

connection.connect(async (err) => {
	if (err) {
		console.error("error connecting: " + err.stack);
		return;
	}

	console.log("connected as id " + connection.threadId);

	connection.query("show tables", function (error, results, fields) {
		if (error) throw error;
		// connected!

		console.log("results:", results);

		let has_configuration_table;

		results.map((row) => {
			if (Object.values(row)[0] === "Configurations") {
				has_configuration_table = true;
			}
		});

		if (has_configuration_table) {
			console.log("select CreateAt, Active, Value from Configurations order by CreateAt");
			connection.query("select CreateAt, Active, Value from Configurations order by CreateAt", async (error, results, fields) => {
				let initial_config, current_config;

				if (error) throw error;
				// connected!

				try {
					results.map((row) => {
						row.CreateAt = new Date(row.CreateAt);

						if (!initial_config) {
							initial_config = row;
						}

						if (row.Active === 1) {
							current_config = row;
						}
					});

					initial_config = JSON.parse(initial_config.Value);
					current_config = JSON.parse(current_config.Value);

					console.log("results.length:", results.length);
					console.log("initial_config:", initial_config);
					console.log("current_config:", current_config);

					compare_and_patch_config(new_config, initial_config, current_config);

					fs.writeFileSync(program.mergedConfig, JSON.stringify(current_config, " ", 4));
				} catch (e) {
					console.log("e:", e);
				}

				console.log("updated current_config:", JSON.stringify(current_config, " ", 4));

				await setTimeout(() => {}, 2000);

				process.exit(0);
			});
		} else {
			process.exit(0);
		}
	});
});

function compare_and_patch_config(new_config, initial_config, current_config) {
	for (prop in new_config) {
		console.log("\n**** Processing prop:", prop);
		console.log("new_config[" + prop + "]:", new_config[prop]);
		console.log("initial_config[" + prop + "]:", initial_config[prop]);
		console.log("current_config[" + prop + "]:", current_config[prop]);

		if (typeof new_config[prop] === "object" && !(new_config[prop] instanceof Array) && new_config[prop] !== null) {
			compare_and_patch_config(new_config[prop], initial_config[prop], current_config[prop]);
			continue;
		}

		if (initial_config[prop] === undefined && current_config[prop] === undefined && new_config[prop] !== undefined) {
			console.log("--- prop", prop, "has NOT been changed by the user");
			console.log("--- prop", prop, 'was CHANGED in the template. CHANGING!"', current_config[prop], '" --> "', new_config[prop], '"');
			current_config[prop] = new_config[prop];
			continue;
		}

		let changed_by_user = current_config[prop] !== initial_config[prop];

		if (changed_by_user) {
			console.log("--- prop", prop, 'was CHANGED by the user. NOT TOUCHING! "', initial_config[prop], '" --> "', current_config[prop], '"');
			continue;
		}

		if (current_config[prop] instanceof Array) {
			changed_by_user = current_config[prop].join("") !== initial_config[prop].join("");
		} else if (typeof current_config[prop] !== "object") {
			changed_by_user = current_config[prop] !== initial_config[prop];
		}

		console.log("changed_by_user:", changed_by_user);

		if (changed_by_user) {
			console.log("--- prop", prop, 'was CHANGED by the user. NOT TOUCHING! "', initial_config[prop], '" --> "', current_config[prop], '"');
		} else {
			console.log("--- prop", prop, "has NOT been changed by the user");

			let changed_in_template = current_config[prop] !== new_config[prop];

			if (new_config[prop] instanceof Array) {
				changed_in_template = new_config[prop].join("") !== current_config[prop].join("");
			}

			if (changed_in_template) {
				console.log("--- prop", prop, 'was CHANGED in the template. CHANGING!"', current_config[prop], '" --> "', new_config[prop], '"');
				current_config[prop] = new_config[prop];
			} else {
				console.log("--- prop", prop, "was NOT changed in the template.");
			}
		}
	}
}
