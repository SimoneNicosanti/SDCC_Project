# import customtkinter as ctk
# import tkinter.messagebox as tkmb
# from controller import Controller

# ctk.set_appearance_mode("dark")
# ctk.set_default_color_theme("blue")
# ctk.set_widget_scaling(1.5)
# ctk.set_window_scaling(1.5)

# loginOnce = 0
# app = ctk.CTk()


# def okButtonAction(action, fileName):
#     actions = {"get": Controller.getFile, "put": Controller.putFile, "delete": Controller.deleteFile}
#     actions[action](fileName)


# def askForFileName(action):
#     frame = ctk.CTkToplevel()
#     frame.title("SAE - Search for Files")
#     title_font = ctk.CTkFont(family="Courier", size=16, slant="italic")
#     label1 = ctk.CTkLabel(frame, text="Please insert file name below:", font=title_font, justify="left")
#     label1.grid(row=0, column=0, columnspan=1, pady=5, padx=15, sticky="w")
#     fileName_entry = ctk.CTkEntry(master=frame, placeholder_text="FileName",
#                                   font=ctk.CTkFont(family="Courier", size=16, slant="italic"), width=250, height=20)
#     fileName_entry.grid(row=1, column=0, pady=24, padx=10, sticky="we")
#     okButton = ctk.CTkButton(master=frame, text='✓', command=lambda: okButtonAction(action, fileName_entry.get()),
#                              fg_color="#ffd000", text_color="black", border_color="black", hover_color="#b99700",
#                              font=("Courier", 16))
#     frame.bind("<Return>", lambda event: okButtonAction(action, fileName_entry.get()))
#     okButton.grid(row=1, column=1, pady=24, padx=10)
#     frame.attributes('-topmost', 'true')


# def administration_interface():
#     global loginOnce, app
#     loginOnce = 1
#     app.destroy()
#     app = ctk.CTk()
#     app.geometry("400x500")
#     app.title("SAE - Administration")
#     for r in range(3):
#         app.grid_rowconfigure(r, weight=1)
#     for c in range(4):
#         app.grid_columnconfigure(c, weight=1)

#     title_font = ctk.CTkFont(family="Courier", size=32, weight="bold")
#     label1 = ctk.CTkLabel(app, text="ADMINISTRATION", font=title_font, justify="left", text_color="#ffd000")
#     label1.grid(row=0, column=0, columnspan=4, pady=5, padx=5, sticky="we")

#     frame = ctk.CTkFrame(master=app)
#     frame.grid(row=2, column=0, columnspan=4, padx=10, pady=15, sticky="we")

#     button_data = [("Get a File", "get"), ("Put a File", "put"), ("Delete a File", "delete")]

#     for i, (text, action) in enumerate(button_data):
#         button = ctk.CTkButton(master=frame, text=text, command=lambda action=action: askForFileName(action),
#                                fg_color="#ffffff", text_color="black", border_color="black", hover_color="#ffd000",
#                                font=("Courier", 16))
#         button.grid(row=i, column=0, pady=24, padx=25, sticky="we")

#     backButton = ctk.CTkButton(master=frame, text='Back to Login', command=startup_login_interface,
#                                fg_color="#ff0000", text_color="white", border_color="black", hover_color="#a00000",
#                                font=("Courier", 16))
#     app.bind("<Escape>", startup_login_interface)
#     backButton.grid(row=len(button_data), column=1, pady=24, padx=25, sticky="we")

#     app.mainloop()


# def login_action(username, password, email):
#     if Controller.login(username, password, email):
#         tkmb.showinfo(title="Login Successful", message="You have logged in Successfully")
#         administration_interface()
#     else:
#         tkmb.showerror(title="Login Failed", message="Invalid Username, password or email")


# def startup_login_interface(event=None):
#     global loginOnce, app
#     if loginOnce:
#         app.destroy()
#         app = ctk.CTk()
#     app.geometry("400x500")
#     app.title("SAE - Login")
#     app.grid_rowconfigure(0, weight=1)
#     app.grid_columnconfigure(0, weight=1)
#     app.grid_columnconfigure(2, weight=1)

#     title_font = ctk.CTkFont(family="Courier", size=16, slant="italic")
#     label1 = ctk.CTkLabel(app, text="""
#       ███████╗ █████╗ ███████╗
#       ██╔════╝██╔══██╗██╔════╝
#       ███████╗███████║█████╗  
#       ╚════██║██╔══██║██╔══╝  
#       ███████║██║  ██║███████╗
#       ╚══════╝╚═╝  ╚═╝╚══════╝
              
#     Storage nel Cloud Continuum
#     """, font=title_font, justify="left", text_color="#ffd000")
#     label1.grid(row=0, column=0, columnspan=1, pady=5)

#     frame = ctk.CTkFrame(master=app)
#     frame.grid(row=2, column=0, columnspan=5, padx=10, pady=15)
#     entry_font = ctk.CTkFont(family="Courier", size=16, slant="italic")

#     entry_labels = ["Username", "Password", "E-Mail"]
#     entry_objects = []

#     for i, label_text in enumerate(entry_labels):
#         if label_text == "Password":
#             entry = ctk.CTkEntry(master=frame, placeholder_text=label_text, font=entry_font, width=250, height=20,
#                                  show="*")
#         else:
#             entry = ctk.CTkEntry(master=frame, placeholder_text=label_text, font=entry_font, width=250, height=20)
#         entry.grid(row=i, column=0, pady=24, padx=10)
#         entry_objects.append(entry)

#     button = ctk.CTkButton(master=frame,
#                            text='Login',
#                            command=lambda: login_action(*[entry.get() for entry in entry_objects]),
#                            fg_color="#ffd000", text_color="black", border_color="black", hover_color="#b99700",
#                            font=("Courier", 16))
#     app.bind("<Return>", lambda event: login_action(*[entry.get() for entry in entry_objects]))
#     app.bind("<Escape>", lambda event: exit())
#     button.grid(row=len(entry_labels), column=0, pady=24, padx=10)

#     app.mainloop()
