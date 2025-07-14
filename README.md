# The-Turing-Project

The Turing Test is a test of a machine's ability to exhibit intelligent behavior that is indistinguishable from that of a human.

It was proposed in 1950 by Alan Turing, a British mathematician and pioneer in computer science, in his paper "Computing Machinery and Intelligence." Instead of asking “Can machines think?”, Turing suggested a more practical question: Can a machine imitate a human well enough to fool another human?

In a world where new large language models are being released even faster than React frameworks, each outperforming the other on some obscure benchmarks set by these AI companies, The Turing Project is an initiative towards **Public Generated Benchmarking (PG-Benchmarking)**, i.e benchmarking done by the public - fully transparent.

The benchmarking criteria is 
> A model's ability to converse like a human

By asking it a fixed set of questions, can we determine if it is a human or an AI based on their responses? Moreover how is the performance of different models (GPT-4, GPT-4o, Claude3, Llama3, Gemini1.5, and many more) on this specific metric.


### The Game
This repository contains a simple Turing-Test game. 
- There are 2 possible modes - you can either play as the **Questioner** (the person asking the questions) or the **Answerer** (the person replying to the questions)
- If you are playing as the Questioner, the game will randomly assign your opponent to be either an LLM or some other human playing as an Answerer, waiting in the queue
- If you are playing as the Answerer, you will be put in a waiting queue. Only when the game has found a opponent, will you be matched with them
- Once both the parties have matched, a peer-to-peer chat session between them will open up where they can talk

#### Game flow for Questioner
![Flow of game for Questioner](https://private-user-images.githubusercontent.com/72456458/447634372-e46e9e79-5881-498b-94ed-d65268e4e5a8.svg?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3NTI0NzkxNTYsIm5iZiI6MTc1MjQ3ODg1NiwicGF0aCI6Ii83MjQ1NjQ1OC80NDc2MzQzNzItZTQ2ZTllNzktNTg4MS00OThiLTk0ZWQtZDY1MjY4ZTRlNWE4LnN2Zz9YLUFtei1BbGdvcml0aG09QVdTNC1ITUFDLVNIQTI1NiZYLUFtei1DcmVkZW50aWFsPUFLSUFWQ09EWUxTQTUzUFFLNFpBJTJGMjAyNTA3MTQlMkZ1cy1lYXN0LTElMkZzMyUyRmF3czRfcmVxdWVzdCZYLUFtei1EYXRlPTIwMjUwNzE0VDA3NDA1NlomWC1BbXotRXhwaXJlcz0zMDAmWC1BbXotU2lnbmF0dXJlPWQ3Y2RlYjM1YjIwYjA1YTQ0ZGJhZjhiMmFhODhkYjQwZWM3ODI3NDI3MDExZjkwOTY2MGQ4ZDVlMjA1MzE4MjQmWC1BbXotU2lnbmVkSGVhZGVycz1ob3N0In0.04vsSrrlfMYOFSujORKqfN1nbXxgPxsqfpMr1Ww3hGw "Game flow for Questioner")

### How to Win?
- As a Questioner, your objective is to correctly determine whether your opponent is a human or a machine. You only get to send 10 messages
- As an Answerer, your objective is to fool your opponent into thinking that you are a machine.

### LLM Models
I have used the Ollama server for the LLM models. For changing which model you want to use (or) modifying the system prompt check out the file `config.yaml`

### Requirements for running the game
- A running Redis server
- A running Ollama server
- Ensure correct server uri's are added in `config.yaml`

### Screens

<img width="1906" height="1071" alt="image (2)" src="https://github.com/user-attachments/assets/1e9da946-3bd6-4072-b34a-62e11beb44f3" />

![image (3)](https://github.com/user-attachments/assets/53067b91-852c-4e2f-a855-1d1ce80ff1e7)

### Note
This project is still in WIP!  
I have quite an ambitious vision for this project and a lot of functionalities have not yet been implemented.  
But keep an eye out :)
