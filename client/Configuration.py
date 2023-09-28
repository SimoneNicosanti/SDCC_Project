import os, jproperties, time


def setUpEnvironment(propertiesFile : str):
	config = jproperties.Properties()
	with open(propertiesFile, "rb") as properties_file:
		config.load(properties_file)
		for key, value in config.items():
			os.environ[key] = value.data

def addAwsCredentialsToEnv():
	with open("/.aws/credentials", "r") as credentials_file:
		credentials_file.readline()
		for i in range(0, 3):
			line = credentials_file.readline()[:-1].split('=')
			os.environ[line[0].upper()] = line[1]