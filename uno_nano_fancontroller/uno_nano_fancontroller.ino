


void setupTimer() {
  //Set PWM frequency to about 25khz on pins 9,10 (timer 1 mode 10, no prescale, count to 320)
  TCCR1A = (1 << COM1A1) | (1 << COM1B1) | (1 << WGM11);
  TCCR1B = (1 << CS10) | (1 << WGM13);
  ICR1 = 320;
  OCR1A = 0;
  OCR1B = 0;
}

//equivalent of analogWrite on pin 9
void setPWM9(float f) {
  f = f < 0 ? 0 : f > 1 ? 1 : f;
  OCR1A = (uint16_t)(320 * f);
}
//equivalent of analogWrite on pin 10
void setPWM10(float f) {
  f = f < 0 ? 0 : f > 1 ? 1 : f;
  OCR1B = (uint16_t)(320 * f);
}

unsigned long volatile tachtimepin2current = 0, tachtimepin2previous = 0;//, tachcountpin2=0;
unsigned long volatile tachtimepin3current = 0, tachtimepin3previous = 0;//, tachcountpin3=0;

byte buff[12];
//Interrupt handler. Stores the timestamps of the last 2 interrupts and handles debouncing

void RisingDifferents2() {
//tachcountpin2++;
  
  tachtimepin2previous = tachtimepin2current;
  tachtimepin2current = micros();
}
void RisingDifferents3() {
 // tachcountpin3++;
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
  return 30000000/difference; //rmp
}
unsigned long calcRPM3() {
  unsigned long difference = tachtimepin3current - tachtimepin3previous;
return 30000000/difference; //rmp
}
void setup() {
  pinMode(2, INPUT);
  pinMode(3, INPUT);
  pinMode(9, OUTPUT);
  pinMode(10, OUTPUT);
  setupTimer();
  attachInterrupt(digitalPinToInterrupt(2), RisingDifferents2, RISING);
  attachInterrupt(digitalPinToInterrupt(3), RisingDifferents3, RISING);
  setPWM9(.4f);
  setPWM10(.4f);
  Serial.begin(19200);  //enable serial so we can see the RPM in the serial monitor

}
void loop() {
  
 
  if(Serial.available()){
    Serial.readBytes(buff,2);
    switch (buff[0]){
    case 0:
     setPWM9((float)(buff[1]));
    break;
    case 1:
      setPWM10((float)(buff[1]));
    break;
    case 2:
      Serial.println(calcRPM2());
    break;
    case 3:
      Serial.println(calcRPM3());
    break;
    default:
    break;
  }
  }
}
