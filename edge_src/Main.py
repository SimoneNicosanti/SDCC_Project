
def main() :
    with open("/temp_edge/prova.txt", "+x") as prova_file :
        prova_file.write("Ciao")

if __name__ == "__main__" :
    main()