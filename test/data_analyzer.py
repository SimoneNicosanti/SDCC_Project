import pandas as pd
import matplotlib.pyplot as plt



def main() :
    sequential_data_analyze()

def sequential_data_analyze() :
    dataframe : pd.DataFrame = pd.read_csv("../docker/Results/Sequential.csv")
    
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
                label = "#Edges = " + str(edgeNum)
            )
            figure.set_xlabel("FileSize [MB]")
            figure.set_ylabel("Time [s]")
        
        plt.legend()
        plt.title(operation)
        plt.savefig("../docker/Results/Charts/Sequential/" + operation)

        plt.clf()
    return


if __name__ == "__main__" :
    main()