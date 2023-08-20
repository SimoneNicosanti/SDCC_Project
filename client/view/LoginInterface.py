import customtkinter as ctk
import tkinter.messagebox as tkmb

ctk.set_appearance_mode("dark")
ctk.set_default_color_theme("blue")

ctk.set_widget_scaling(1.5)
ctk.set_window_scaling(1.5)

loginOnce = 0
app = ctk.CTk()

def create_new_window():
    global loginOnce
    loginOnce = 1
    global app  # Access the global variable to destroy the old window
    app.destroy()  # Close and delete the old window
    # Create a new main window
    app = ctk.CTk()
    app.geometry("400x500")
    app.title("ACCESS SUCCESSFUL")
    app.resizable(False, False)
    app.grid_rowconfigure(0, weight=1)
    app.grid_rowconfigure(2, weight=1)
    app.grid_columnconfigure(0, weight=1)
    app.grid_columnconfigure(3, weight=1)
    # Set up the new window's contents
    label = ctk.CTkLabel(app, text="New Window")
    label.pack()
    button = ctk.CTkButton(master=app, text='Back to Login', command=base,
                    fg_color="yellow",
                    text_color="black",
                    border_color="black",
                    hover_color="orange")
    button.pack()
    # ... (other widgets and setup for the new window)
    app.mainloop()
    
    # ... (other widgets and setup for the new window)

def login(username, password, email, event=None):
    tkmb.showinfo(title="Login Successful",message="You have logged in Successfully")
    create_new_window()
    #tkmb.showerror(title="Login Failed",message="Invalid Username, password and email")


def base():
    global loginOnce
    global app
    if loginOnce:
        app.destroy()  # Close and delete the old window
        # Create a new main window
        app = ctk.CTk()
    app.geometry("400x500")
    app.title("SC2 - LOGIN")
    app.resizable(False, False)
    app.grid_rowconfigure(0, weight=1)
    app.grid_rowconfigure(2, weight=1)
    app.grid_columnconfigure(0, weight=1)
    app.grid_columnconfigure(3, weight=1)
    label1 = ctk.CTkLabel(app, text=
    """
    ███████╗ ██████╗██████╗ 
    ██╔════╝██╔════╝╚════██╗
    ███████╗██║      █████╔╝
    ╚════██║██║     ██╔═══╝ 
    ███████║╚██████╗███████╗
    ╚══════╝ ╚═════╝╚══════╝
    """,
        font=("Courier", 16),
        justify="left"
    )
    label1.grid(row=0, column=0, columnspan=1)  # Span both columns

    my_font = ctk.CTkFont(family="Courier", size=20, slant="italic")

    label2 = ctk.CTkLabel(app, text=
    """
    Storage nel Cloud Continuum
    """,
        font=my_font,
        justify="left"
    )
    label2.grid(row=1, column=0, columnspan=1)  # Span both columns

    frame = ctk.CTkFrame(master=app)
    frame.grid(row=2, column=0, columnspan=5, padx=10, pady=10)
    entry_font = ctk.CTkFont(family="Courier", size=16, slant="italic")

    user_entry = ctk.CTkEntry(master=frame, placeholder_text="Username", font=entry_font, width=250, height=20)
    user_entry.grid(row=0, column=0, pady=12, padx=10, sticky="w")

    user_pass = ctk.CTkEntry(master=frame, placeholder_text="Password", show="*", font=entry_font, width=250, height=20)
    user_pass.grid(row=1, column=0, pady=12, padx=10, sticky="w")

    user_email = ctk.CTkEntry(master=frame, placeholder_text="E-Mail", font=entry_font, width=250, height=20)
    user_email.grid(row=2, column=0, pady=12, padx=10, sticky="w")

    button = ctk.CTkButton(master=frame, 
                           text='Login', 
                           command=lambda: login(user_entry.get(), user_pass.get(), user_email.get()),
                           fg_color="yellow",
                           text_color="black",
                           border_color="black",
                           hover_color="orange")
    app.bind("<Return>", lambda event: login(user_entry.get(), user_pass.get(), user_email.get()))
    button.grid(row=3, column=0, pady=12, padx=10)

    app.mainloop()

if __name__ == "__main__" :
    base()