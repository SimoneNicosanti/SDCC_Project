
from view import OptionsInterface
import Configuration as Configuration

if __name__ == "__main__":
    Configuration.setUpEnvironment("conf.properties")
    Configuration.addAwsCredentialsToEnv()
    OptionsInterface.main()

    