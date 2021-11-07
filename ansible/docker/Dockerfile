FROM python:3
#This is based on Debian 
#Installing oh my zsh to mke life easier

WORKDIR /usr/src/app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
RUN apt-get update && apt-get install -y nano zsh curl && sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
    
COPY . .

CMD [ "/usr/bin/zsh" ]