import pandas as pd
import matplotlib.pyplot as plt



def main() :
    sequential_data_analyze()
    # parallel_data_analyze()

def parallel_data_analyze() :
    line_plot("Parallel")
    return

def sequential_data_analyze() :
    line_plot("Sequential")
    return


def line_plot(fileName : str) :
    dataframe : pd.DataFrame = pd.read_csv("../docker/Results/" + fileName + ".csv")
    
    renamedDataframe = dataframe.rename(columns = {"FileSize [MB]" : "FileSize", "Time [s]" : "Time"})
    # renamedDataframe = renamedDataframe[renamedDataframe.FileSize < 100]

    avgDataframe = renamedDataframe.groupby(["RequestType", "FileSize", "EdgeNum"]).aggregate("mean").reset_index()

    ## Per ogni operazione traccio un grafico
    for operation in avgDataframe["RequestType"].unique() :
        operationDataFrame : pd.DataFrame = avgDataframe[avgDataframe.RequestType == operation]
        ## Per ogni edgeNum traccio una curva
        figure = plt.subplot()
        for edgeNum in operationDataFrame["EdgeNum"].unique() :
            chartDataFrame : pd.DataFrame = operationDataFrame[operationDataFrame.EdgeNum == edgeNum]
            figure.plot(
                chartDataFrame["FileSize"], 
                chartDataFrame["Time"], 
                label = "#Edges = " + str(edgeNum),
                marker = "o"
            )
            figure.set_xlabel("FileSize [MB]")
            figure.set_ylabel("Time [s]")
        
        plt.legend()
        plt.title(operation)
        plt.grid()
        plt.savefig("../docker/Results/Charts/" + fileName + "/" + operation)

        plt.clf()

        figure = plt.subplot()
        for edgeNum in operationDataFrame["EdgeNum"].unique() :
            chartDataFrame : pd.DataFrame = operationDataFrame[(operationDataFrame.EdgeNum == edgeNum) & (operationDataFrame.FileSize <= 30)]
            figure.plot(
                chartDataFrame["FileSize"], 
                chartDataFrame["Time"], 
                label = "#Edges = " + str(edgeNum),
                marker = "o"
            )
            figure.set_xlabel("FileSize [MB]")
            figure.set_ylabel("Time [s]")
        
        plt.legend()
        plt.title(operation)
        plt.grid()
        plt.savefig("../docker/Results/Charts/" + fileName + "/" + operation + "_CUT")
        plt.clf()
    return




if __name__ == "__main__" :
    main()