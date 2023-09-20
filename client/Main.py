#from view import LoginInterface
from controller import Controller
from view import OptionsInterface
import Configuration as Configuration

if __name__ == "__main__":
    Configuration.setUpEnvironment("conf.properties")
    OptionsInterface.main()

    