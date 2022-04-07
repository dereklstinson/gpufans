


void setupTimer() {
  //Set PWM frequency to about 25khz on pins 9,10 (timer 1 mode 10, no prescale, count to 320)
  TCCR1A = (1 << COM1A1) | (1 << COM1B1) | (1 << WGM11);
  TCCR1B = (1 << CS10) | (1 << WGM13);
  ICR1 = 320;
  OCR1A = 0;
  OCR1B = 0;
}

//equivalent of analogWrite on pin 9
void setPWM9(long int b) {
  long int x = (320 * b);
  long int y = (999);
  OCR1A = (uint16_t)(x / y);
}
//equivalent of analogWrite on pin 10
void setPWM10(long int b) {
  long int x = (320 * b);
  long int y = (999);
  OCR1B = (uint16_t)(x / y);
}

unsigned long volatile tachtimepin2current = 0, tachtimepin2previous = 0, tachcountpin2 = 0;
unsigned long volatile tachtimepin3current = 0, tachtimepin3previous = 0, tachcountpin3 = 0;
unsigned long volatile globaltime = 0;
//Interrupt handler. Stores the timestamps of the last 2 interrupts and handles debouncing

void RisingDifferents2() {
  tachcountpin2++;
  tachtimepin2previous = tachtimepin2current;
  tachtimepin2current = micros();
}
void RisingDifferents3() {
  tachcountpin3++;
  tachtimepin3previous = tachtimepin3current;
  tachtimepin3current = micros();
}
//Calculates the RPM based on the timestamps of the last 2 interrupts. Can be called at any time.
unsigned long calcRPM2() {

  unsigned long difference = tachtimepin2current - tachtimepin2previous;
  //two rises per rotation
  //60 seconds in a minute
  //1,000,000 microseconds in a second
  //60000000ms/minute * 1 rotation/(2*difference(microseconds))
  //30000000/difference
  return 30000000 / difference; //rpm
}
unsigned long calcRPM3() {
  unsigned long difference = tachtimepin3current - tachtimepin3previous;
  return 30000000 / difference; //rpm

}

void setup() {
  pinMode(2, INPUT);
  pinMode(3, INPUT);
  pinMode(9, OUTPUT);
  pinMode(10, OUTPUT);
  setupTimer();

  attachInterrupt(digitalPinToInterrupt(2), RisingDifferents2, RISING); //Fan 0
  attachInterrupt(digitalPinToInterrupt(3), RisingDifferents3, RISING);  //Fan 1
  setPWM9(300); //Fan 0
  setPWM10(300); //Fan 1
  Serial.begin(19200);  //enable serial so we can see the RPM in the serial monitor

}
char buff[12];
void loop() {
  unsigned long nowtime = millis();
  if ( nowtime < globaltime) {
    globaltime = nowtime;
    tachcountpin2 = 0;
    tachcountpin3 = 0;

  }
  unsigned long elapsedtime = nowtime - globaltime;
  unsigned long rpm2 = 1;
  unsigned long rpm3 = 1;
  if ( elapsedtime >= 5000) {
    rpm2 = tachcountpin2 / elapsedtime;
    rpm3 = tachcountpin3 / elapsedtime;
    tachcountpin2 = 0;
    tachcountpin3 = 0;
    globaltime = nowtime;


  }

  if (Serial.available()) {
    Serial.readBytes(buff, 4);
    String st10 = String(buff[1]);
    String st20 = String(buff[2]);
    String st30 = String(buff[3]);
    String st00 = String(st10 + st20 + st30);
    Serial.flush();
    switch (buff[0]) {
      case '0':
        setPWM9(st00.toInt());
        break;
      case '1':
        setPWM10(st00.toInt());
        break;
      case '2':
        if (rpm2 == 0) {
          Serial.println("0000");
        } else {
          Serial.println(calcRPM2());
        }

        break;
      case '3':
        if (rpm3 == 0) {
          Serial.println("0000");
        } else {
          Serial.println(calcRPM3());
        }

        break;
      default:
        break;
    }
  }

}
