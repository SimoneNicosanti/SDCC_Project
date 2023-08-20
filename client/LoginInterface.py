import customtkinter as ctk
import tkinter.messagebox as tkmb
from controller import Controller

ctk.set_appearance_mode("dark")
ctk.set_default_color_theme("blue")

ctk.set_widget_scaling(1.5)
ctk.set_window_scaling(1.5)

loginOnce = 0
app = ctk.CTk()

def okButtonAction(action, fileName):
    if action == "get":
        Controller.getFile(fileName)
    elif action == "put":
        Controller.putFile(fileName)
    else :
        Controller.deleteFile(fileName)

def askForFileName(action):
    frame = ctk.CTkToplevel()
    title_font = ctk.CTkFont(family="Courier", size=16, slant="italic")
    label1 = ctk.CTkLabel(frame, text="Please insert file name below:",font=title_font,justify="left")
    label1.grid(row=0, column=0, columnspan=1, pady=5)
    fileName_entry = ctk.CTkEntry(master=frame, placeholder_text="FileName", font=ctk.CTkFont(family="Courier", size=16, slant="italic"), width=250, height=20)
    fileName_entry.grid(row=1, column=0, pady=24, padx=10)
    okButton = ctk.CTkButton(master=app, text='Confirm', command=lambda: okButtonAction(action, fileName_entry.get()),
                    fg_color="#ffd000",
                    text_color="black",
                    border_color="black",
                    hover_color="#b99700",
                    font=("Courier", 16))
    frame.bind("<Return>", lambda event: okButtonAction(fileName_entry.get()))
    okButton.grid(row=2, column=0, pady=24, padx=10)

def administration_interface():
    global loginOnce
    loginOnce = 1
    global app  # Access the global variable to quit the old window
    app.destroy()  # Close and delete the old window
    # Create a new main window
    app = ctk.CTk()
    app.geometry("400x500")
    app.title("SAE - ADMINISTRATION")
    app.resizable(False, False)
    app.grid_rowconfigure(0, weight=1)
    app.grid_rowconfigure(2, weight=1)
    app.grid_columnconfigure(0, weight=1)
    app.grid_columnconfigure(3, weight=1)
    # Set up the new window's contents
    title_font = ctk.CTkFont(family="Courier", size=16, slant="italic")
    label1 = ctk.CTkLabel(app, text=
    """
      ███████╗ █████╗ ███████╗
      ██╔════╝██╔══██╗██╔════╝
      ███████╗███████║█████╗  
      ╚════██║██╔══██║██╔══╝  
      ███████║██║  ██║███████╗
      ╚══════╝╚═╝  ╚═╝╚══════╝
              
    Welcome back, choose one option
    """,
        font=title_font,
        justify="left"
    )
    label1.grid(row=0, column=0, columnspan=1, pady=5)  # Span both columns

    frame = ctk.CTkFrame(master=app)
    frame.grid(row=2, column=0, columnspan=5, padx=10, pady=15)

    getButton = ctk.CTkButton(master=app, text='Get a File', command=lambda:askForFileName("get"),
                    fg_color="#ffd000",
                    text_color="black",
                    border_color="black",
                    hover_color="#b99700",
                    font=("Courier", 16))
    getButton.grid(row=0, column=0, pady=24, padx=10)

    putButton = ctk.CTkButton(master=app, text='Put a File', command=lambda:askForFileName("put"),
                    fg_color="#ffd000",
                    text_color="black",
                    border_color="black",
                    hover_color="#b99700",
                    font=("Courier", 16))
    putButton.grid(row=1, column=0, pady=24, padx=10)

    deleteButton = ctk.CTkButton(master=app, text='Delete a File', command=lambda:askForFileName("delete"),
                    fg_color="#ffd000",
                    text_color="black",
                    border_color="black",
                    hover_color="#b99700",
                    font=("Courier", 16))
    deleteButton.grid(row=2, column=0, pady=24, padx=10)

    backButton = ctk.CTkButton(master=app, text='Back to Login', command=startup_login_interface,
                    fg_color="#ff0000",
                    text_color="white",
                    border_color="black",
                    hover_color="#a00000",
                    font=("Courier", 16))
    app.bind("<Escape>", startup_login_interface)
    backButton.grid(row=3, column=0, pady=24, padx=10)

    app.mainloop()

def login_action(username, password, email):
    if Controller.login(username, password, email):
        tkmb.showinfo(title="Login Successful",message="You have logged in Successfully")
        administration_interface()
    else:
        tkmb.showerror(title="Login Failed",message="Invalid Username, password or email")


def startup_login_interface():
    global loginOnce
    global app
    if loginOnce:
        app.destroy()  # Close and delete the old window
        # Create a new main window
        app = ctk.CTk()
    app.geometry("400x500")
    app.title("SAE - LOGIN")
    app.resizable(False, False)
    app.grid_rowconfigure(0, weight=1)
    app.grid_columnconfigure(0, weight=1)
    app.grid_columnconfigure(2, weight=1)

    title_font = ctk.CTkFont(family="Courier", size=16, slant="italic")
    label1 = ctk.CTkLabel(app, text=
    """
      ███████╗ █████╗ ███████╗
      ██╔════╝██╔══██╗██╔════╝
      ███████╗███████║█████╗  
      ╚════██║██╔══██║██╔══╝  
      ███████║██║  ██║███████╗
      ╚══════╝╚═╝  ╚═╝╚══════╝
              
    Storage nel Cloud Continuum
    """,
        font=title_font,
        justify="left"
    )
    label1.grid(row=0, column=0, columnspan=1, pady=5)  # Span both columns

    frame = ctk.CTkFrame(master=app)
    frame.grid(row=2, column=0, columnspan=5, padx=10, pady=15)
    entry_font = ctk.CTkFont(family="Courier", size=16, slant="italic")

    user_entry = ctk.CTkEntry(master=frame, placeholder_text="Username", font=entry_font, width=250, height=20)
    user_entry.grid(row=0, column=0, pady=24, padx=10)

    user_pass = ctk.CTkEntry(master=frame, placeholder_text="Password", show="*", font=entry_font, width=250, height=20)
    user_pass.grid(row=1, column=0, pady=24, padx=10)

    user_email = ctk.CTkEntry(master=frame, placeholder_text="E-Mail", font=entry_font, width=250, height=20)
    user_email.grid(row=2, column=0, pady=24, padx=10)

    button = ctk.CTkButton(master=frame, 
                           text='Login', 
                           command=lambda: login_action(user_entry.get(), user_pass.get(), user_email.get()),
                           fg_color="#ffd000",
                           text_color="black",
                           border_color="black",
                           hover_color="#b99700",
                           font=("Courier", 16))
    app.bind("<Return>", lambda event: login_action(user_entry.get(), user_pass.get(), user_email.get()))
    button.grid(row=3, column=0, pady=24, padx=10)

    app.mainloop()

if __name__ == "__main__" :
    startup_login_interface()