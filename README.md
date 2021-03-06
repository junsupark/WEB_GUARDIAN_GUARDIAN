# GUARDIAN - CCTV감지체계

# Logo
![Logo](https://github.com/osamhack2020/WEB_GUARDIAN_GUARDIAN/blob/main/logo.png)

## 팀소개 및 프로잭트 설명 동영상
[추가 예정]  

## 기능 설계
**[Oven](https://ovenapp.io/project/yx0NjBwiaWnct4suJmMu9fvgJmmRNLwe#sy1Vs)**  

## 컴퓨터 구성 / 필수 조건 안내 (Prerequisites)
* ECMAScript 6 지원 브라우저 사용
* 권장: Google Chrome 버젼 77 이상

## 기술 스택 (Technique Used)
### Server(back-end)
 - golang 1.15.2
 - GoCV (OpenCV Binding) 
 - mongoDB
 - Express
 
### front-end
 - TypeScript
 - React.js  
 - Antd UI Framework  
 - Socket.io

### Infra
 - Docker

### Machine-learning
 - Darknet Yolov4-tiny
 
## 설치 안내 (Installation Process)
- NodeJS + yarn 과 git 설치 (리눅스 우분투 기준)
```bash
$ sudo apt install git
$ sudo apt install nodejs
$ npm install yarn -g
```
- 깃허브에서 프로젝트 받아오기(클론)
```bash
$ git clone https://github.com/osamhack2020/WEB_GUARDIAN_GUARDIAN
```
- 노드모듈 설치
```bash
$ yarn add global nodemon
$ yarn install
```
- 실행(프론트엔드)
```bash
$ yarn start
```
## 프로젝트 사용법 (Getting Started)
- 백엔드 NodeJS서버 사용법
```bash
$ cd backend
$ yarn upgrade
$ node index.js 나 nodemon index.js
// 노드몬을 쓰면 파일이변경될때마다 재시작해줌. 개발용
```
인터넷 브라우저 접속 http://localhost

## 팀 정보 (Team Information)
- 강선규 (gron1gh1@gmail.com), Github Id: gron1gh1
- 신기철 (skck0226@gmail.com), Github Id: skck0226
- 장성호 (csi9725@naver.com), Github Id: JangSeongHo99
- 백승민 (zlws554@naver.com), Github Id: me9min
- 이승규 (dltmdrb98@gmail.com), Github Id: seungy0
- 오택환 (jamesoh3928@naver.com), Github Id:jamesoh3928

## 저작권 및 사용권 정보 (Copyleft / End User License)
 * [MIT](https://github.com/osamhack2020/WEB_GUARDIAN_GUARDIAN/blob/main/LICENSE)
