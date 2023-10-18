
from view import OptionsInterface
import Configuration as Configuration
from performance_test import TestRunner
import sys



if __name__ == "__main__":
    Configuration.setUpEnvironment("conf.properties")
    Configuration.addAwsCredentialsToEnv()
    if (len(sys.argv) == 1) :
        OptionsInterface.main()
    else :
        TestRunner.RunTest(sys.argv)


