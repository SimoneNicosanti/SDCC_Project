import os
import jproperties

	

def setUpEnvironment(propertiesFile : str):

	config = jproperties.Properties()

	with open(propertiesFile, "rb") as properties_file :
		config.load(properties_file)
		for key, value in config.items():
			os.environ[key] = value.data