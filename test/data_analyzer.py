import pandas as pd
import matplotlib.pyplot as plt



def main() :
    sequential_data_analyze()
    parallel_data_analyze()

def parallel_data_analyze() :
    line_plot("Parallel")
    return

def sequential_data_analyze() :
    line_plot("Sequential")
    return


def line_plot(fileName : str) :
    dataframe : pd.DataFrame = pd.read_csv("./Results/" + fileName + ".csv")
    
    renamedDataframe = dataframe.rename(columns = {"FileSize [MB]" : "FileSize", "Time [s]" : "Time"})
    # renamedDataframe = renamedDataframe[renamedDataframe.FileSize < 100]

    avgDataframe = renamedDataframe.groupby(["RequestType", "FileSize", "EdgeNum"]).aggregate("mean").reset_index()

    ## Per ogni operazione traccio un grafico
    for operation in avgDataframe["RequestType"].unique() :
        operationDataFrame : pd.DataFrame = avgDataframe[avgDataframe.RequestType == operation]
        ## Per ogni edgeNum traccio una curva
        figure = plt.subplot()
        figure.set_facecolor('#292626')
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
            figure.tick_params(axis='x', colors='#C6EAC9')
            figure.tick_params(axis='y', colors='#C6EAC9')
            figure.xaxis.label.set_color('#C6EAC9')
            figure.yaxis.label.set_color('#C6EAC9')
        plt.legend()
        plt.title(operation).set_color('#C6EAC9')
        plt.grid()
        plt.savefig("./Results/Charts/" + fileName + "/" + operation, transparent=True)

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
            figure.tick_params(axis='x', colors='#C6EAC9')
            figure.tick_params(axis='y', colors='#C6EAC9')
            figure.xaxis.label.set_color('#C6EAC9')
            figure.yaxis.label.set_color('#C6EAC9')
        
        plt.legend()
        plt.title(operation + "_CUT").set_color('#C6EAC9')
        plt.grid()
        plt.savefig("./Results/Charts/" + fileName + "/" + operation + "_CUT", transparent=True)
        plt.clf()
    return

def box_plot(fileName : str) :
    dataframe : pd.DataFrame = pd.read_csv("./Results/" + fileName + ".csv")
    
    renamedDataFrame = dataframe.rename(columns = {"FileSize [MB]" : "FileSize", "Time [s]" : "Time"})
    # renamedDataframe = renamedDataframe[renamedDataframe.FileSize < 100]

    ## Per ogni operazione traccio un grafico
    for operation in renamedDataFrame["RequestType"].unique() :
        fileSizeList = renamedDataFrame.FileSize.unique()
        edgeNumList = renamedDataFrame["EdgeNum"].unique()
        fileSizeList.sort()
        edgeNumList.sort()

        operationDataFrame : pd.DataFrame = renamedDataFrame[renamedDataFrame.RequestType == operation]

        ## Per ogni edgeNum traccio una curva
        figure, axs = plt.subplots(nrows = edgeNumList.size, ncols = fileSizeList.size, sharex = True)
        for i in range(0, edgeNumList.size) :
            edgeNum = edgeNumList[i]
            for j in range(0, fileSizeList.size) :
                fileSize = fileSizeList[j]

                dataFrame = operationDataFrame[(operationDataFrame.FileSize == fileSize) & (operationDataFrame.EdgeNum == edgeNum)]
                axs[i][j].boxplot(dataFrame["Time"], widths = 0.3)
                axs[i][j].set_title("Edge = " + str(edgeNum) + "\n" + "Size = " + str(fileSize))
                axs[i][j].xaxis.set_visible(False)
                axs[i][j].grid()
        
        figure.set_size_inches(9, 9)
        
        figure.suptitle(operation)
        figure.subplots_adjust(wspace = 0, hspace = 0)
        figure.tight_layout()
        figure.savefig("./Results/Charts/" + fileName + "/" + operation + "_BOX")
        figure.clf()
        plt.clf()
    return


if __name__ == "__main__" :
    main()