import os
import jproperties

def set_properties_as_env_variables(config):
	for key, value in config.items():
	    os.environ[key] = value.data

def main():
	filename = "conf.properties"

	properties_config = jproperties.Properties()

	with open(filename, "rb") as properties_file :
		properties_config.load(properties_file)

		set_properties_as_env_variables(properties_config)